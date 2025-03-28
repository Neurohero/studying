Playbook запускает установку контейнеров Grafana + Prometheus и node_exporter на ВМ1, так же устанавливает контейнеры с node и postgres_exporter-ами на 2 и 3ВМ, конфигурационный файл prometheus включает джобы для сбора метрик с хостов с экспортерами (подключается как volume). 

Два compose файла: 
compose_gf устанавливает образы grafana + prometheus, node_exporter на 1ВМ
compose_np устанавливает образы node_exporter и postgres_exporter на 2 и 3ВМ