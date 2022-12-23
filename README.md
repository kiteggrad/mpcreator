# mpcreator

Основная цель проекта - предоставить простой способ массово склонировать / актуализировать (git pull) / ... репозитории из gitlab.

- Позволяет склонировать в одну директорию все репозитории из gitlab. На выходе получается git репозиторий с сабмодулями.
  - сохраняется иерархия gitlab групп (каждой группе соответствует папка)
  - если репозиторий уже существует он никак не изменяется
- Позволяет стянуть последние изменения (like git pull) во все репозитории одной командой
  - если репозиторий не на основной ветке - он никак не изменяется

Запустить можно путём `go run main.go` или сбилженного бинарника `go install github.com/kiteggrad/mpcreator@latest` (изучить --help)

## Examples
- Создание и клонирование всех интересующих репозиториев
  
  ```bash
  # mkdir my-company
  # cd my-company

  # тянем всё
  mpcreator fill -p . -u ${GITLAB_URL} -t ${GITLAB_TOKEN}

  # или только для интересующей нас группы / репозитория / языка
  mpcreator fill -p . -u ${GITLAB_URL} -t ${GITLAB_TOKEN} --ingroups "some-group" --inprojects "some1,some2" --inlang "Go"
  ```
  При этом папка `my-company` - может уже существовать и содержать my-company/some-group. Ничего страшного не произойдёт, ничего внутри репозитория задето не будет. 
  
  Если архитектура групп и проектов в gitlab отличается от существующей файловой в `my-company` - возникнут дубли репозиториев, например: `my-company/some-group/some1`, `my-company/some1`

- Получение последних изменений для склонированных репозиториев
  
  ```bash
  # cd my-company

  # тянем всё
  mpcreator pull -p . -u ${GITLAB_URL} -t ${GITLAB_TOKEN} --ingroups "some-group"

  # или только для интересующей нас группы / репозитория
  mpcreator pull -p . -u ${GITLAB_URL} -t ${GITLAB_TOKEN} --ingroups "some-group" --inprojects "some1,some2"
  ```
  При этом если в обновляемом репозитории текущая ветка отличается от основной - ничего не произойдёт (выведется WARN лог).

- В примерах перечислены не все аргументы для `mpcreator` / `mpcreator fill` / `mpcreator pull` / ... читайте --help для каждой команды.

- Рекомендуется так-же положить в `my-company` Makefile похожего содержания
    ```Makefile
    include .env
    export

    deps:
        go install github.com/kiteggrad/mpcreator@latest

    fill-some-group-go:
        mpcreator fill -p . -u ${GITLAB_URL} -t ${GITLAB_TOKEN} --ingroups "some-group" --inlang "Go"

    pull-some-group:
        mpcreator pull -p . -u ${GITLAB_URL} -t ${GITLAB_TOKEN} --ingroups "some-group"
    ```

## TODO
- (?) Флаг для извлечения только реп к которым есть определённый уровень доступа
- autocomplete

---
<details>
    <summary>Drafts</summary>
На написанное ниже можно не обращать внимание, просто черновики

Рекомендуется к прочтению https://git-scm.com/book/ru/v2/Инструменты-Git-Подмодули

Команды для обновления сабмодулей:
- вливает отслеживаемую ветку (по умолчанию master), даже если сейчас не на ней.
`git submodule update --init --recursive --remote --merge`
- подтягивает изменения только если сабмодуль на отслеживаемой ветке, не валится если произошла ошибка в одном из сабмодулей.
`git submodule foreach "git pull origin --recurse-submodules --ff-only || true"`
</details>
