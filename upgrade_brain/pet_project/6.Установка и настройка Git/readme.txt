Используя композник установить gitlab, после этого выполнить команду sudo docker exec -it gitlab grep 'Password:' /etc/gitlab/initial_root_password 
для получения пароля администратора. После установить runner используя композник, после установки раннера, подключиться в контейнер и выполнить
gitlab-runner register  --url https://gitlab.mylittledev.ru  --token указать_токен, токен взять из gitlab https://gitlab.mylittledev.ru/admin/runners/new при создании 
runner-а
