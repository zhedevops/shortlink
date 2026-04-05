# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

### Результат оптимизации:
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

File: shortener

Build ID: f1187321500581560d909d4fa6f269c9cd2ce625

Type: cpu

Time: 2026-04-04 16:03:10 MSK

Duration: 17.16s, Total samples = 160ms ( 0.93%)

Showing nodes accounting for -70ms, 43.75% of 160ms total

flat  flat%   sum%        cum   cum%

0     0% 43.75%      -10ms  6.25%  github.com/zhedevops/shortlink/internal/handler.(*Handler).CreateShortLinkHandler

0     0% 43.75%      -10ms  6.25%  github.com/zhedevops/shortlink/internal/handler.(*Handler).GetLinkByIDHandler

0     0% 43.75%      -20ms 12.50%  github.com/zhedevops/shortlink/internal/middleware.GzipHandle.func1

0     0% 43.75%      -20ms 12.50%  github.com/zhedevops/shortlink/internal/middleware.Logger.func1

0     0% 43.75%      -10ms  6.25%  github.com/zhedevops/shortlink/internal/model.NewShortys

0     0% 43.75%      -10ms  6.25%  github.com/zhedevops/shortlink/internal/router.NewRouter.RequireContentType.func1.1

0     0% 43.75%      -10ms  6.25%  github.com/zhedevops/shortlink/internal/service.(*Service).GetOriginalURL

0     0% 43.75%      -10ms  6.25%  github.com/zhedevops/shortlink/internal/storage.(*DBStorage).GetOriginalURL
