### 0. Предварительные требования

В качестве первого шага мы должны следовать главной рекомендации при работе с программным обеспечением — проверка на наличие обновлений и установка последних пакетов:

```shell
sudo apt update
sudo apt upgrade
```

Кроме того, необходимо загрузить и установить несколько зависимостей:
```shell
sudo apt install software-properties-common apt-transport-https ca-certificates gnupg2 gpg curl
```

### 1. Отключение swap

Мы можем узнать, использует ли наша машина swap-память, используя команду `htop`:
![[Pasted image 20250127010627.png]]

Мы можем отключить swap, выполнив следующую команду:

```shell
swapoff -a
```

Чтобы убедиться, что swap остается отключенным после запуска, нам необходимо закомментировать строку в файле `/etc/fstab`, который инициализирует swap-память при загрузке Linux:
![[Pasted image 20250127010727.png]]В этом конкретном случае в качестве раздела swap использовался файл с именем `swap.img`, удаляем его:

```shell
sudo rm /swap.img 
```

### 2. Включение модулей ядра Linux

Прежде чем продолжить, мы войдем под учетную запись root и выполним все нижеуказанные действия с привилегиями этого суперпользователя:

```shell
sudo su - root
```
Вот два модуля ядра Linux, которые нам нужно включить:

1. **br_netfilter** — этот модуль необходим для включения прозрачного маскирования и облегчения передачи трафика [Virtual Extensible LAN (VxLAN)](https://habr.com/ru/articles/344326/) для связи между Kubernetes Pods в кластере.
    
2. **overlay** — этот модуль обеспечивает необходимую поддержку на уровне ядра для правильной работы драйвера хранения overlay. По умолчанию модуль overlay может не быть включен в некоторых дистрибутивах Linux, и поэтому необходимо включить его вручную перед запуском Kubernetes.
    

Мы можем включить эти модули, выполнив команду `modprobe` вместе с флагом `-v`(подробный вывод), чтобы увидеть результат:

```shell
modprobe overlay -v
modprobe br_netfilter -v
```

После этого мы должны получить следующий вывод:![[Pasted image 20250127011206.png]]

Чтобы убедиться, что модули ядра загружаются после перезагрузки, мы также можем добавить их в файл `/etc/modules`:

```shell
echo "overlay" >> /etc/modules
echo "br_netfilter" >> /etc/modules
```

После включения модуля `br_netfilter` необходимо включить IP-пересылку в ядре Linux, чтобы обеспечить сетевое взаимодействие между подами и узлами. IP-пересылка позволяет ядру Linux маршрутизировать пакеты с одного сетевого интерфейса на другой.

Для этого мы должны записать "1" в файл конфигурации с названием "`ip_forward`":

```shell
echo 1 > /proc/sys/net/ipv4/ip_forward
```

### 3. Установка Kubelet

Установка Kubelet, возможно, самый простой шаг, поскольку он очень хорошо описан в [официальной документации Kubernetes](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl). В основном, нам нужно выполнить следующие команды:

```shell
# If the directory `/etc/apt/keyrings` does not exist, it should be created before the curl command.
# sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.32/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
```

```shell
# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.32/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
```

```shell
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable --now kubelet
```


### 4. Установка среды выполнения контейнеров

```shell
curl -fsSL https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_22.04/Release.key | sudo gpg --dearmor -o /usr/share/keyrings/libcontainers-archive-keyring.gpg
curl -fsSL https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/1.28/xUbuntu_22.04/Release.key | sudo gpg --dearmor -o /usr/share/keyrings/libcontainers-crio-archive-keyring.gpg
```

``` shell
echo "deb [signed-by=/usr/share/keyrings/libcontainers-archive-keyring.gpg] https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_22.04/ /" | sudo tee /etc/apt/sources.list.d/libcontainers.list
echo "deb [signed-by=/usr/share/keyrings/libcontainers-crio-archive-keyring.gpg] https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/1.28/xUbuntu_22.04/ /" | sudo tee /etc/apt/sources.list.d/crio.list
```

```shell
sudo apt update 
sudo apt install -y cri-o cri-o-runc
sudo systemctl enable crio
sudo systemctl start crio
```

Команда `systemctl status crio` должна вывести текущее состояние службы:
![[Pasted image 20250127012457.png]]

Проверить, что контейнеры запускаются:

```shell
sudo crictl ps
```
### 5. Инициализация мастера
```shell
sudo kubeadm init --apiserver-advertise-address=192.168.0.100 --pod-network-cidr=10.10.0.0/16 --dry-run
```

Прежде чем применять наши изменения, мы можем запустить настройку сети с использованием флага `--dry-run` без каких-либо изменений.

Если так называемые "предварительные проверки" выдают ошибку, то, выполнив быстрый поиск в Google, мы можем исправить проблему и в последствии применить эти изменения без флага `--dry-run`

Необходимо скопировать учетные данные администратора, которые будут использоваться `kubectl` для управления ресурсами нашего кластера через компонент API-сервера Kubernetes:

``` shell
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
```

``` shell
kubectl get pods
```
### 6. Инициализация кластера

После того, как мы получили виртуальную машину, которая проходит предварительные проверки, мы можем инициализировать наш кластер.

После выполнения этой команды kubeadm превратит нашу виртуальную машину в [узел управления Kubernetes](https://kubernetes.io/docs/concepts/overview/components/), состоящий из следующих основных компонентов:

- **etcd** — хранилище ключ-значение, используемое для хранения состояния всего кластера Kubernetes;
    
- **kube-scheduler** — компонент управления, который отслеживает вновь созданные Pod без назначенного узла и выбирает для них узел для запуска;
    
- **kube-controller-manager** — компонент управления, запускающий процессы контроллера.

![[Pasted image 20250127012733.png]]

### 7. Добавление воркера

Предварительно на worker ноду требуется установить kubectl, kubeadm, kubelet, так же провести настройку виртуальной машины, для этого необходимо выполнить пункты с 1-4, затем выполнить команду присоединения

``` shell
kubeadm join 192.168.0.100:6443 --token nljqps.vypo4u9y07lsw7s2 \        --discovery-token-ca-cert-hash sha256:f820767cfac10cca95cb7649569671a53a2240e1b91fcd12ebf1ca30c095c2d6
```
