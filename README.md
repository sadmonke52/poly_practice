# poly_practise_1_R

## Описание

Этот проект демонстрирует монолитную архитектуру на Go с использованием Kafka для обмена сообщениями, Prometheus для сбора метрик и Grafana для визуализации. Включает генератор, продюсер, консьюмер и агрегатор сообщений.

## Основные технологии

- Go (Golang)

- Docker, docker-compose

- Prometheus (мониторинг)

- Grafana (визуализация)

- Собственная система логирования

## Быстрый старт

1. Запуск инфраструктуры

Запустите Kafka, Redis, Prometheus и Grafana через Docker Compose:

   docker-compose up -d

1. Запуск приложения
    
    
       go run cmd/MonolithicArchitecture/main.go
    

2. Доступ к сервисам и метрикам

- Kafka: localhost:9092

- Kafka UI: localhost:8080

- Redis: localhost:6379

- Prometheus: http://localhost:9090

- Grafana: http://localhost:3000

## Конфигурация

- Основные параметры настраиваются в config/local.yaml.

- Метрики доступны по адресу /metrics 


# Poly Practise 2 R

## Описание

Этот проект состоит из нескольких микросервисов, написанных на Go, и предназначен для демонстрации или практики работы с потоками данных, агрегацией, логированием и метриками. В проекте используются Docker и docker-compose для оркестрации сервисов, а также Prometheus и Grafana для мониторинга.

## Основные технологии

- Go (Golang)

- Docker, docker-compose

- Prometheus (мониторинг)

- Grafana (визуализация)

- Собственная система логирования

## Как запустить


       docker-compose up --build
    

 Сервисы будут доступны в контейнерах, мониторинг — через Prometheus и Grafana.

Доступ к сервисам и метрикам

1) Kafka: localhost:9092

2) Kafka UI: localhost:8080

3) Redis: localhost:6379

4) Prometheus: http://localhost:9090

5) Grafana: http://localhost:3000

## Конфигурация

- Все сервисы используют свои YAML-файлы для локальной конфигурации.

- Метрики и логи собираются и доступны для мониторинга.

## Мониторинг

- Prometheus собирает метрики со всех сервисов.

- Grafana использует дашборды из папки provisioning/dashboards/.
