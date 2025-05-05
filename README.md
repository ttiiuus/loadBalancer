<div align="center">
  <a href="https://github.com/ttiiuus/loadBalancer">
    <img src="https://i.imgur.com/xyz123.gif" width="800" alt="Load Balancer Demo"/>
    <br>
    <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&size=26&duration=2500&pause=1000&color=8A2BE2&center=true&width=800&lines=🚀+Next-gen+Load+Balancer;⚡+Powered+by+Go;🔁+Intelligent+Traffic+Routing" alt="Animated title"/>
  </a>
</div>
# loadBalancer
Round-robin http load balancer

#Реализовано
Round-Robin
Rate-limit
Graceful shutdown
logger


Простой балансировщик на основе Round Robin
#Запуск
git clone https://github.com/ttiiuus/loadBalancer.git
cd loadBalancer
go build -o loadbalancer main.go

#Запуск бинарника в контейнере Docker
git clone https://github.com/ttiiuus/loadBalancer.git
cd loadBalancer
docker build -t loadbalancer .
docker run -p 8080:8080 -v ./configs:/app/configs loadbalancer

#Example 
В конфиге необходимо отредактировать адреса бэкендов и их максимальное кол-во подключений

TODO
Добавить конфиг в yaml
Реализовать дополнительные алгоритмы балансировки + балансировка round-robin c весами
