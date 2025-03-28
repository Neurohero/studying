								УСТАНОВКА POSTGRESQL

Playbook устанавливает Postgresql15 и Postgresql15-server на CentOS 8 и 9, предварительно необходимо загрузить репозитрии с официального сайта Postgres, с данными ОС есть трудности, репозиториев для них нет, в данном случае использовались унифицированные репозитории для CentOS

Ссылки на репозитории и порядок установки для CentOS8:
sudo dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-x86_64/pgdg-redhat-repo-latest.noarch.rpm

Ссылки на репозитории и порядок установки для CentOS9:
sudo dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-9-x86_64/pgdg-redhat-repo-latest.noarch.rpm

Предварительно посмотреть архитектуру процессора через uname -a, в случае если архтитектура не x86_64, то пройти на https://download.postgresql.org/pub/repos/yum/reporpms/ и выбрать актуальную с актуальным репозиторием


==============================================================================================================================================

								СОЗДАНИЕ БД, СХЕМЫ И ТАБЛИЦ
1) Создать базу данных:

CREATE DATABASE db_name;

2) Создать схему:

CREATE SCHEMA test;


3) Создать таблицы:

CREATE TABLE test.test1 (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    familiya VARCHAR(100),
    otchestvo VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

4) Скрипт на автоматическое заполнение таблиц лежит /home/kmendeleev/db_script.sh

============================================================================================================================================

								НАСТРОЙКА ЛОГИЧЕСКОЙ РЕПЛИКАЦИИ:

В папке с postgre (/var/lib/pgsql/15/data) требуется внести изменения в файлы: postgresql.conf и pg_hba.conf на мастер сервере

1) postgresql.conf:

Раскомментировать или написать следующие параметры:

wal_level = logical
max_replication_slots = 4
max_wal_senders = 4

2) pg_hba.conf:

Дописать хост для подключения:

host    replication     all         10.10.10.10/32        md5
host    test    postgres    10.10.10.10/32    md5

3) Создать публикацию:

CREATE PUBLICATION test FOR TABLE test.test1, test.test2, test.test3, test.test4, test.test5; 

Чтобы добавить новую таблицу в публикацию используйте:

ALTER PUBLICATION test ADD TABLE test.table100,

Выполнить рестарт сервиса


4) На реплицируемом сервере требуется произвести аналогичные настройки указав хост для подключения мастер сервера в pg_hba.conf, за исключением пункта 3.

На реплицируемом сервере создается подписка на публикацию с реплицирующего сервера:

CREATE SUBSCRIPTION test
CONNECTION 'host=10.255.254.102 port=5432 user=postgres password=qQ345zpHxn# dbname=test'
PUBLICATION test;

Выполнить рестарт сервиса
