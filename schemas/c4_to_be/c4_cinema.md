# Допущения
1. Контент реплицирован и стримится агрегатором
Партнёр:
- передаёт видеоконтент (файлы, исходники, mezzanine)
-  метаданные
- условия лицензии (регионы, срок, тип доступа)
Агрегатор:
- хранит видео у себя (S3 / Object Storage)
- кодирует (HLS/DASH, разные битрейты)
- раздаёт через свой CDN
- контролирует DRM, токены, сессии
При воспроизведении:
- никакого запроса к серверам партнёра
- всё происходит внутри инфраструктуры агрегатора
 
2. Доступен только просмотр фильмов по единой подписке кинобездны (без возможности покупки отдельных фильмов и без возможности покупки подписки отдельного контент-провайдера). 

```puml

@startuml
!includeurl https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml
LAYOUT_WITH_LEGEND()
top to bottom direction

skinparam defaultFontSize 9
skinparam wrapWidth 160
skinparam Padding 2
skinparam NodePadding 2
skinparam ArrowThickness 0.7

title Кинобездна — C4 Container Diagram 

Person(webUser, "Пользователь (Web)")
Person(mobileUser, "Пользователь (Mobile)")
Person(tvUser, "Пользователь (Smart TV)")


System_Ext(extProviders, "Контент-провайдеры", "Передача метаданных + условий лицензии")
System_Ext(extReco, "Внешняя рекомендательная система", "Обмен через RabbitMQ (async)")
System_Ext(extPSP, "Платёжные провайдеры", "PSP")


System_Boundary(s1, "Кинобездна") {

  Boundary(edge, "Edge") {
    Container(apiGw, "API Gateway", "Kong", "Единая точка входа: auth, rate limiting, routing")

  }

  Boundary(bff, "BFF") {
    Container(bffWeb, "BFF Web", "Go", "Композиция ответов под Web")
    Container(bffMobile, "BFF Mobile", "Go", "Композиция ответов под Mobile")
    Container(bffTV, "BFF TV", "Go", "Композиция ответов под Smart TV")
  }

  Container(kafka, "Kafka (Internal Event Bus)", "Kafka", "Внутренние доменные события (async)")

  Boundary(core, "Core Services") {
    Container(userSvc, "User Service", "Go", "Профиль/настройки пользователя")
    Container(reviewsSvc, "Reviews Service", "Go", "Оценки/отзывы пользователей")
    Container(recoSvc, "Recommendations Service", "Go", "Получение и обработка рекомендаций через RabbitMQ")
    Container(entSvc, "Entitlements Service", "Go", "Решение доступа к контенту")
    Container(playbackSvc, "Playback Service", "Go", "Старт и условия просмотра")
    Container(providerSvc, "Providers Service", "Go", "Лицензии/окна доступности")
    Container(storage, "Content Storage + CDN", "CDN", "Хранение, кодирование, раздача через CDN")
    Container(moviesSvc, "Movies Service (Catalog & Metadata)", "Go", "Единый каталог и метаданные")
  }

  Container(rmqReco, "RabbitMQ (Reco Integration)", "AMQP", "Интеграция с внешней рекомендательной системой")

  Boundary(commerce, "Commerce") {
    Container(orchestratorSvc, "Orchestrator Service", "Go", "Оркестрация оплаты подписки")
    Container(subsSvc, "Subscriptions Service", "Go", "Информация о подписках")
    Container(paymentsSvc, "Payments Service", "Go", "Интеграция с PSP, статусы платежей, webhooks")
    Container(promosSvc, "Promotions Service", "Go", "Промокоды/скидки")  
  }
}

' ---------- Client entry ----------
Rel(webUser, apiGw, "HTTPS")
Rel(mobileUser, apiGw, "HTTPS")
Rel(tvUser, apiGw, "HTTPS")

Rel(apiGw, bffWeb, "Routes")
Rel(apiGw, bffMobile, "Routes")
Rel(apiGw, bffTV, "Routes")

' ---------- BFF composition ----------
Rel(bff, core, "Reads/writes", "HTTP")
Rel(bff, orchestratorSvc, "Reads/writes", "HTTP")

' ---------- Recommendations: ONLY RabbitMQ with external system ----------
Rel(recoSvc, rmqReco, "Publish request / Consume esponse", "AMQP")
Rel(extReco, rmqReco, "Consume request / Publish response", "AMQP")
Rel(recoSvc, entSvc, "Фильтр доступности выдачи", "HTTP")

' ---------- Playback runtime (без запросов к провайдеру) ----------
Rel(playbackSvc, entSvc, "Проверка доступности просмотра", "HTTP")
Rel(playbackSvc, storage, "Выдача параметров воспроизведения", "HTTP")

' ---------- Provider ingest & licensing (offline/async) ----------
Rel(providerSvc, extProviders, "Получение контента и лицензий", "HTTP")
Rel(providerSvc, storage, "Загрузка и публикация контента", "HTTP")

' Доменные события из core(Kafka)
Rel(core, kafka, "Publish events", "Events")
Rel(kafka, core, "Consume events", "Events")

' ---------- Commerce ----------
Rel(orchestratorSvc, subsSvc, "Команды управления подпиской", "HTTP")
Rel(orchestratorSvc, paymentsSvc, "Команды проведения платежа", "HTTP")
Rel(subsSvc, promosSvc, "Проверка применимости скидок", "HTTP")
Rel(paymentsSvc, extPSP, "Проведение платежа и вебхуки", "HTTPS")

' Доменные события из commerce (Kafka)
Rel(commerce, kafka, "Publish events", "Events")

@enduml


```
