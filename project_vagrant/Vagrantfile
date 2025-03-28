# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.ssh.insert_key = false
  config.vm.boot_timeout = 600
  config.vm.provision "file", source: "vagrant.private.key.pem", destination: "~/.ssh/id_rsa"

  config.vm.define 'machine120' do |cs|
    cs.vm.box = "bento/ubuntu-20.04"
#	config.vm.box_url = "file://#{Dir.pwd}/ubuntu-20.04.box"
    cs.vm.hostname = 'machine120.test'
    cs.vm.network 'private_network', ip: '192.168.56.120'
	
	  # Проброс порта Grafana на localhost хоста
    cs.vm.network "forwarded_port",
      guest: 3000,
      host: 3000,
      host_ip: "127.0.0.1",
      auto_correct: true,
      protocol: "tcp"
  
    cs.vm.network "forwarded_port", guest: 9090, host: 9090, host_ip: "127.0.0.1" # Prometheus
    cs.vm.network "forwarded_port", guest: 9100, host: 9100, host_ip: "127.0.0.1" # Node Exporter

    cs.vm.provision "shell", inline: <<-SHELL
      # Обновление и установка Ansible
      sudo apt-get update
      sudo apt-get install -y software-properties-common
      sudo apt-add-repository -y ppa:ansible/ansible
      sudo apt-get update
      sudo apt-get install -y ansible
	  sudo apt-get install -y docker.io docker-compose curl
      sudo usermod -aG docker vagrant
    SHELL

    # Настройка Ansible provisioner
    cs.vm.provision "ansible_local" do |ansible|
      ansible.playbook = "playbooks/120.yml"
    end

    # Настройки VirtualBox
    cs.vm.provider 'virtualbox' do |v|
      v.memory = 4096
      v.cpus = 2
      v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
      v.customize ["modifyvm", :id, "--ioapic", "on"]
    end
  end
end