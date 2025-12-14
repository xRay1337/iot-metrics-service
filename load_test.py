#!/usr/bin/env python3
"""
Нагрузочное тестирование IoT Metrics Service
Цель: 1000 RPS в течение 5 минут
"""

import asyncio
import aiohttp
import random
import time
import json
import statistics
from datetime import datetime
import argparse
import sys

class LoadTester:
    def __init__(self, base_url, target_rps=1000, duration=300, device_count=100):
        self.base_url = base_url
        self.target_rps = target_rps
        self.duration = duration
        self.device_count = device_count
        self.results = {
            'success': 0,
            'errors': 0,
            'latencies': [],
            'anomalies': 0
        }
        self.start_time = None
        self.session = None
        
    def generate_metric(self, device_id=None):
        """Генерация случайной метрики IoT устройства"""
        if device_id is None:
            device_id = f"device-{random.randint(1, self.device_count):03d}"
        
        # Создаем более реалистичные паттерны нагрузки
        hour = datetime.now().hour
        time_factor = 1.0
        
        # Имитация суточных колебаний
        if 2 <= hour <= 5:  # Ночь
            time_factor = 0.3 + random.random() * 0.3
        elif 9 <= hour <= 17:  # Рабочий день
            time_factor = 0.8 + random.random() * 0.4
        elif 18 <= hour <= 22:  # Вечер
            time_factor = 0.6 + random.random() * 0.4
        else:  # Утро/полночь
            time_factor = 0.4 + random.random() * 0.4
        
        # Иногда создаем аномалии (5% случаев)
        is_anomaly = random.random() < 0.05
        
        if is_anomaly:
            cpu = random.uniform(90, 99)  # Аномально высокая нагрузка
        else:
            cpu = random.uniform(10, 80) * time_factor
            
        return {
            "timestamp": int(time.time()),
            "device_id": device_id,
            "cpu": round(cpu, 2),
            "rps": random.uniform(100, 2000) * time_factor,
            "memory": random.uniform(30, 90) * time_factor
        }
    
    async def send_request(self, session, request_num):
        """Отправка одного запроса"""
        try:
            metric = self.generate_metric()
            start = time.time()
            
            async with session.post(
                f"{self.base_url}/api/metrics",
                json=metric,
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                latency = time.time() - start
                
                if response.status == 202:
                    self.results['success'] += 1
                    self.results['latencies'].append(latency * 1000)  # в мс
                    
                    # Проверяем, была ли это аномалия
                    if metric['cpu'] > 90:
                        self.results['anomalies'] += 1
                        
                    return True, latency
                else:
                    self.results['errors'] += 1
                    print(f"[{request_num}] HTTP {response.status}")
                    return False, latency
                    
        except asyncio.TimeoutError:
            self.results['errors'] += 1
            print(f"[{request_num}] Timeout")
            return False, 5.0
        except Exception as e:
            self.results['errors'] += 1
            print(f"[{request_num}] Exception: {e}")
            return False, 5.0
    
    async def monitor_progress(self):
        """Мониторинг прогресса теста"""
        while True:
            elapsed = time.time() - self.start_time
            if elapsed >= self.duration:
                break
                
            remaining = self.duration - elapsed
            success_rate = (self.results['success'] / 
                          (self.results['success'] + self.results['errors']) * 100 
                          if (self.results['success'] + self.results['errors']) > 0 else 0)
            
            current_rps = self.results['success'] / elapsed if elapsed > 0 else 0
            
            print(f"\r⏱️  {elapsed:.1f}s / {self.duration}s | "
                  f"RPS: {current_rps:.1f} | "
                  f"Success: {self.results['success']} | "
                  f"Errors: {self.results['errors']} | "
                  f"Rate: {success_rate:.1f}% | "
                  f"Anomalies: {self.results['anomalies']}", end="", flush=True)
            
            await asyncio.sleep(1)
    
    async def run_test(self):
        """Запуск нагрузочного теста"""
        print("🚀 IoT Metrics Service - Load Test")
        print("=" * 60)
        print(f"Target URL: {self.base_url}")
        print(f"Target RPS: {self.target_rps}")
        print(f"Duration: {self.duration}s ({self.duration/60:.1f} min)")
        print(f"Devices: {self.device_count}")
        print("=" * 60)
        
        # Проверка доступности сервиса
        print("🔍 Checking service availability...")
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.base_url}/api/health", timeout=5) as resp:
                    if resp.status == 200:
                        health = await resp.json()
                        print(f"✅ Service is {health.get('status', 'unknown')}")
                        print(f"   Redis: {health.get('redis', 'unknown')}")
                    else:
                        print(f"❌ Service unavailable: HTTP {resp.status}")
                        return
        except Exception as e:
            print(f"❌ Cannot connect to service: {e}")
            return
        
        # Запуск теста
        print("\n⚡ Starting load test...")
        self.start_time = time.time()
        
        connector = aiohttp.TCPConnector(limit=0, limit_per_host=0)
        async with aiohttp.ClientSession(connector=connector) as session:
            self.session = session
            
            # Запускаем мониторинг прогресса
            monitor_task = asyncio.create_task(self.monitor_progress())
            
            # Основной цикл отправки запросов
            request_num = 0
            while time.time() - self.start_time < self.duration:
                tasks = []
                
                # Создаем пакет запросов
                batch_size = min(self.target_rps // 4, 50)  # Отправляем пакетами
                for _ in range(batch_size):
                    request_num += 1
                    tasks.append(self.send_request(session, request_num))
                
                # Отправляем пакет
                await asyncio.gather(*tasks)
                
                # Регулируем скорость отправки
                elapsed_batch = time.time() - self.start_time
                target_requests = self.target_rps * elapsed_batch
                actual_requests = self.results['success'] + self.results['errors']
                
                if actual_requests < target_requests:
                    # Нужно ускориться
                    await asyncio.sleep(0.01)
                else:
                    # Нужно замедлиться
                    await asyncio.sleep(0.05)
            
            # Завершаем мониторинг
            monitor_task.cancel()
            try:
                await monitor_task
            except asyncio.CancelledError:
                pass
        
        # Выводим итоги
        self.print_summary()
    
    def print_summary(self):
        """Вывод итогов тестирования"""
        total_time = time.time() - self.start_time
        total_requests = self.results['success'] + self.results['errors']
        avg_rps = self.results['success'] / total_time if total_time > 0 else 0
        success_rate = (self.results['success'] / total_requests * 100 
                       if total_requests > 0 else 0)
        
        # Статистика latency
        latencies = self.results['latencies']
        if latencies:
            avg_latency = statistics.mean(latencies)
            p95_latency = statistics.quantiles(latencies, n=20)[18]  # 95 перцентиль
            p99_latency = statistics.quantiles(latencies, n=100)[98]  # 99 перцентиль
        else:
            avg_latency = p95_latency = p99_latency = 0
        
        print("\n" + "=" * 60)
        print("📊 LOAD TEST RESULTS")
        print("=" * 60)
        print(f"Total time:      {total_time:.2f}s")
        print(f"Total requests:  {total_requests}")
        print(f"Successful:      {self.results['success']}")
        print(f"Failed:          {self.results['errors']}")
        print(f"Success rate:    {success_rate:.2f}%")
        print(f"Average RPS:     {avg_rps:.2f}")
        print(f"Target RPS:      {self.target_rps}")
        print(f"Simulated anomalies: {self.results['anomalies']}")
        print("-" * 60)
        print("⏱️  Latency (ms):")
        print(f"  Average:       {avg_latency:.2f}")
        print(f"  95th percentile: {p95_latency:.2f}")
        print(f"  99th percentile: {p99_latency:.2f}")
        print("-" * 60)
        
        # Проверяем достижение целей
        if avg_rps >= self.target_rps * 0.9:
            print("✅ TARGET RPS ACHIEVED!")
        else:
            print(f"⚠️  Target RPS not achieved: {avg_rps:.2f} < {self.target_rps}")
        
        if avg_latency < 100:  # < 100ms
            print("✅ LATENCY WITHIN LIMITS")
        else:
            print(f"⚠️  High latency: {avg_latency:.2f}ms")
        
        if success_rate >= 95:
            print("✅ HIGH SUCCESS RATE")
        else:
            print(f"⚠️  Low success rate: {success_rate:.2f}%")
        
        print("=" * 60)
        
        # Дополнительные проверки
        print("\n🔍 Additional checks:")
        try:
            import requests
            # Проверяем аналитику
            resp = requests.get(f"{self.base_url}/api/analyze?device_id=device-001", timeout=5)
            if resp.status_code == 200:
                data = resp.json()
                print(f"✅ Analytics working (device-001 avg: {data.get('rolling_average', 0):.2f})")
            
            # Проверяем аномалии
            resp = requests.get(f"{self.base_url}/api/anomalies", timeout=5)
            if resp.status_code == 200:
                data = resp.json()
                print(f"✅ Anomalies endpoint working (found: {data.get('count', 0)})")
            
            # Проверяем Prometheus метрики
            resp = requests.get(f"{self.base_url}/api/prometheus", timeout=5)
            if resp.status_code == 200:
                print("✅ Prometheus metrics available")
                
        except Exception as e:
            print(f"⚠️  Additional checks failed: {e}")

def main():
    parser = argparse.ArgumentParser(description='Load test IoT Metrics Service')
    parser.add_argument('--url', default='http://localhost:8080',
                       help='Base URL of the service (default: http://localhost:8080)')
    parser.add_argument('--rps', type=int, default=1000,
                       help='Target requests per second (default: 1000)')
    parser.add_argument('--duration', type=int, default=300,
                       help='Test duration in seconds (default: 300 = 5 min)')
    parser.add_argument('--devices', type=int, default=100,
                       help='Number of simulated devices (default: 100)')
    parser.add_argument('--warmup', type=int, default=10,
                       help='Warmup time in seconds (default: 10)')
    
    args = parser.parse_args()
    
    print("📡 IoT Metrics Service - Performance Test")
    print("=" * 60)
    
    # Warm-up
    if args.warmup > 0:
        print(f"🔥 Warming up for {args.warmup}s...")
        time.sleep(args.warmup)
    
    # Запуск теста
    tester = LoadTester(
        base_url=args.url,
        target_rps=args.rps,
        duration=args.duration,
        device_count=args.devices
    )
    
    # Для Windows нужна специальная обработка asyncio
    if sys.platform == 'win32':
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    
    asyncio.run(tester.run_test())

if __name__ == '__main__':
    main()