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
