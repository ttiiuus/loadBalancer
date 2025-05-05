<div align="center">
  <a href="https://github.com/ttiiuus/loadBalancer">
    <img src="https://img.shields.io/badge/⚡-Load_Balancer-8A2BE2?style=for-the-badge&logo=go&logoColor=white" alt="Load Balancer"/>
    <br>
    <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=26&duration=2800&pause=1000&color=8A2BE2&center=true&width=500&lines=🚀+Round-Robin+Balancer;⚡+Powered+by+Go;🛡️+Rate-limiting+%26+Graceful+shutdown" alt="Animated features"/>
  </a>
</div>

## 🌀 О проекте

Высокопроизводительный HTTP балансировщик нагрузки с алгоритмом Round Robin, написанный на Go. Поддерживает ограничение скорости запросов и плавное завершение работы.

## 🌟 Возможности

- 🔁 **Round-Robin** - циклическое распределение запросов
- 🚦 **Rate-limiting** - ограничение частоты запросов
- 🛑 **Graceful shutdown** - плавное завершение работы
- 📝 **Логирование** - детальное логирование операций
- 🐳 **Docker-поддержка** - готовый образ для развертывания

## 🚀 Быстрый старт

### Сборка из исходников
```bash
git clone https://github.com/ttiiuus/loadBalancer.git
cd loadBalancer
go build -o loadbalancer main.go

### Сборка бинарника 
```bash
git clone https://github.com/ttiiuus/loadBalancer.git
cd loadBalancer
go build -o loadbalancer main.go
