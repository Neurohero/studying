#!/bin/bash

# Настройки
SMTP_SERVER="10.128.228.66:25"
EMAIL_FROM="vitrina_egrn_q2@rosreestr.ru"
#EMAIL_TO="k.mendeleev@bft.ru"
EMAIL_TO="vitrina_egrn_q2@rosreestr.ru"
SMTP_AUTH_USER="vitrina_egrn_q2@rosreestr.ru"
SMTP_AUTH_PASS="nLHhw92h%VFVnLHhw92h%VFV222"
PORT=8091

# Переменная для хранения состояния отправки
email_sent=0
last_sent_time=0
send_interval=3600  # Интервал отправки в секундах

# Функция для проверки доступности почты
check_email() {
    echo "Проверка работоспособности почты" | mailx -v -r "$EMAIL_FROM" -s "Проверка почты" \
        -S smtp="$SMTP_SERVER" -S smtp-use-starttls \
        -S smtp-auth=login -S smtp-auth-user="$SMTP_AUTH_USER" \
        -S smtp-auth-password="$SMTP_AUTH_PASS" -S ssl-verify=ignore "$EMAIL_TO"
#        -S smtp-headers="Disposition-Notification-To: $EMAIL_FROM" \
#        "$EMAIL_TO"    # Запрос ответа о доставке
    if [ $? -eq 0 ]; then
        email_sent=1  # Успешная отправка
        echo "$(date): Письмо успешно отправлено" >> /var/log/email_healthcheck.log
    else
        email_sent=0  # Ошибка отправки
        echo "$(date): Ошибка отправки письма" >> /var/log/email_healthcheck.log
    fi
}

# Функция для обработки HTTP-запросов
handle_request() {
    {
        echo -ne "HTTP/1.1 200 OK\r\n"
        echo -ne "Content-Type: text/plain; version=0.0.4; charset=utf-8\r\n"
        echo -ne "\r\n"
        echo -ne "# HELP email_send_success Indicates if email was sent successfully\n"
        echo -ne "# TYPE email_send_success gauge\n"
        echo -ne "email_send_success{instance=\"email_checker\"} $email_sent\n"
    } | nc -l -p "$PORT" -w 1  >> /dev/null 2>&1 # Тайм-аут 1 секунда
}

# Бесконечный цикл для проверки почты и обработки запросов
while true; do
    current_time=$(date +%s)  # Получаем текущее время в секундах

    if (( current_time - last_sent_time >= send_interval )); then
        check_email
        last_sent_time=$current_time  # Обновляем время последней отправки
    fi

    # Запускаем обработчик запросов в фоновом режиме
    handle_request &

    sleep 1  # Проверка каждую секунду
done

