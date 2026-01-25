## Изучите [README.md](README.md) файл и структуру проекта.

## Задание 1

Архитектура to be КиноБездны:
[ссылка на файл](https://github.com/troxanna/architecture-pro-cinemaabyss/blob/cinema/schemas/c4_to_be/c4_cinema.png)


## Задание 2

Результаты тестов:

<img width="524" height="350" alt="Снимок экрана 2026-01-25 в 22 19 08" src="https://github.com/user-attachments/assets/097f9b77-aade-4817-af74-2dce848513e4" />


Состояние топиков Kafka http://localhost:8090:

<img width="1662" height="555" alt="Снимок экрана 2026-01-25 в 22 20 14" src="https://github.com/user-attachments/assets/80a024dd-9163-4268-9b45-c552cd11d9d5" />


## Задание 3

### CI/CD

Результаты сборки:

<img width="1391" height="548" alt="Снимок экрана 2026-01-25 в 22 00 19" src="https://github.com/user-attachments/assets/0e7b731f-9500-4b39-9355-08915ab96d4a" />


Результаты тестов:

<img width="1187" height="552" alt="Снимок экрана 2026-01-25 в 22 08 51" src="https://github.com/user-attachments/assets/13a540c3-5895-445b-a3b4-173ceaa8ad0b" />


### Proxy в Kubernetes

Результат вывода http://cinemaabyss.example.com/api/movies:

<img width="1348" height="252" alt="Снимок экрана 2026-01-25 в 20 03 19" src="https://github.com/user-attachments/assets/14a4091e-d76a-4431-8d20-193d2f4b7d1a" />

Результат вывода event-service после вызова тестов:

<img width="1305" height="216" alt="Снимок экрана 2026-01-25 в 20 35 34" src="https://github.com/user-attachments/assets/e531c281-142a-4d5c-a756-c47c05b828be" />

## Задание 4

Результат развертывания helm:

<img width="1276" height="888" alt="Снимок экрана 2026-01-25 в 21 54 04" src="https://github.com/user-attachments/assets/2e2fd73c-43ff-43ca-b025-4eab75459d50" />

Результат вывода http://cinemaabyss.example.com/api/movies:

<img width="1355" height="455" alt="Снимок экрана 2026-01-25 в 21 55 01" src="https://github.com/user-attachments/assets/58ae57ea-5e59-4586-8326-e4151e46a587" />


# Задание 5

Результат работы circuit breaker'а:

Запросы:
<img width="883" height="118" alt="Снимок экрана 2026-01-26 в 00 45 10" src="https://github.com/user-attachments/assets/7700e31c-f32c-4d7e-94a0-57c818cace28" />

Статистика:
<img width="1344" height="136" alt="Снимок экрана 2026-01-26 в 00 46 08" src="https://github.com/user-attachments/assets/8e4a8952-662b-462b-a9b7-4ac2461f6be5" />

