Написал плейбук и инишник. Плейбук обновляет устанавливает зависимости для докера, добавляет репозиторий докера, устанавливает и запускает докер.

Важно! После спулливания репы перекиньте .ini файл в директорию /etc/ansible/inventory, или ту, которая прописана в файле ansible.cfg.

Запусить плейбук можно командой:

ansible-playbook -i /etc/ansible/inventory/vm.ini job_job/pet_project/1.Установка\ Docker/docker-install.yml -k --ask-become-pass
