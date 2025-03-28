package main

import (
    "crypto/tls"
//  "fmt"
    "io"
    "io/ioutil"
    "log"
    "strings"
    "github.com/PuerkitoBio/goquery"
    "github.com/emersion/go-imap"
    "github.com/emersion/go-imap/client"
    "github.com/emersion/go-message/mail"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "gopkg.in/yaml.v2"
    "net/http"
    "time"
)

type Config struct {
    Server      string `yaml:"server"`
    Login       string `yaml:"login"`
    Password    string `yaml:"password"`
    SenderEmail string `yaml:"sender_email"`
}

var (
    instanceStatus = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "server_status",
            Help: "Status of the servers",
        },
        []string{"instance"},
    )
)

func init() {
    prometheus.MustRegister(instanceStatus)
}

func loadConfig() (*Config, error) {
    data, err := ioutil.ReadFile("config.yaml")
    if err != nil {
        return nil, err
    }

    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}

func main() {
    // Загрузка конфигурации
    config, err := loadConfig()
    if err != nil {
        log.Fatalf("Ошибка загрузки конфигурации: %v", err)
    }

    // Запуск HTTP сервера для экспорта метрик
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Fatal(http.ListenAndServe(":8080", nil))
    }()

    // Основной цикл для обработки писем
    for {
        // Подключение к почтовому серверу
        c, err := client.DialTLS(config.Server, &tls.Config{})
        if err != nil {
            log.Println("Ошибка подключения к почтовому серверу:", err)
            time.Sleep(10 * time.Second) // Ждём перед повторной попыткой
            continue
        }
        defer c.Logout()

        // Вход в систему
        if err := c.Login(config.Login, config.Password); err != nil {
            log.Println("Ошибка входа:", err)
            time.Sleep(10 * time.Second) // Ждём перед повторной попыткой
            continue
        }

        // Открываем папку "INBOX"
        mbox, err := c.Select("INBOX", false)
        if err != nil {
            log.Println("Ошибка открытия папки INBOX:", err)
            continue
        }

        if mbox.Messages == 0 {
            log.Println("Нет сообщений в папке INBOX")
            time.Sleep(10 * time.Second) // Ждём перед следующим запросом
            continue
        }

        // Поиск писем от указанного отправителя
        criteria := imap.NewSearchCriteria()
        criteria.Header.Add("From", config.SenderEmail)

        ids, err := c.Search(criteria)
        if err != nil {
            log.Println("Ошибка поиска писем:", err)
            time.Sleep(10 * time.Second) // Ждём перед повторной попыткой
            continue
        }

        if len(ids) == 0 {
            log.Println("Нет писем от указанного отправителя")
            time.Sleep(10 * time.Second) // Ждём перед следующим запросом
            continue
        }

        seqset := new(imap.SeqSet)
        seqset.AddNum(ids...)

        messages := make(chan *imap.Message, 10)
        err = c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
        if err != nil {
            log.Println("Ошибка извлечения писем:", err)
            time.Sleep(10 * time.Second) // Ждём перед повторной попыткой
            continue
        }

        for msg := range messages {
            for _, r := range msg.Body {
                mr, err := mail.CreateReader(r)
                if err != nil {
                    log.Println("Ошибка создания читателя для письма:", err)
                    continue
                }

                for {
                    part, err := mr.NextPart()
                    if err != nil {
                        break
                    }

                    switch h := part.Header.(type) {
                    case *mail.InlineHeader:
                        contentType, _, _ := h.ContentType()
                        if strings.Contains(contentType, "text/html") {
                            body, err := io.ReadAll(part.Body)
                            if err != nil {
                                log.Println("Ошибка чтения тела письма:", err)
                                continue
                            }

                            // Обработка HTML и обновление метрик
                            updateMetricsFromHTML(string(body))
                        }
                    }
                }
            }
        }

        // Ждём перед следующей итерацией
        time.Sleep(1 * time.Minute) // или установите нужный вам интервал
    }
}

// updateMetricsFromHTML извлекает данные из HTML и обновляет метрики для Prometheus
func updateMetricsFromHTML(htmlContent string) {
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
    if err != nil {
        log.Println("Ошибка парсинга HTML:", err)
        return
    }

    doc.Find("table").Each(func(index int, tableHtml *goquery.Selection) {
        tableHtml.Find("tr").Each(func(rowIndex int, rowHtml *goquery.Selection) {
            var instance, status string

            rowHtml.Find("td").Each(func(colIndex int, cellHtml *goquery.Selection) {
                cellText := strings.TrimSpace(cellHtml.Text())

                if strings.Contains(cellText, "UNX_FS_VITR") || strings.Contains(cellText, "/vitrina/") {
                    instance = cellText
                }
                if cellText == "Successful" || cellText == "Partially Successful" || cellText == "Failed" {
                    status = cellText
                }
            })

            if instance != "" && status != "" {
                instance = strings.TrimPrefix(instance, "UNX_FS_VITR_")
                instance = strings.TrimPrefix(instance, "/vitrina/")

                // Обновляем метрики Prometheus в зависимости от статуса
                switch status {
                case "Successful":
                    instanceStatus.WithLabelValues(instance).Set(1)
                case "Partially Successful":
                    instanceStatus.WithLabelValues(instance).Set(0)
                case "Failed":
                    instanceStatus.WithLabelValues(instance).Set(-1)
                }
            }
        })
    })
}

