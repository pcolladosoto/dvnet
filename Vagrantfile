Vagrant.configure("2") do |config|
    config.vm.box = "ubuntu/jammy64"
    config.ssh.forward_agent = true
    config.ssh.forward_x11 = true

    config.vm.provider "virtualbox" do |v|
        v.customize ["modifyvm", :id, "--memory", 1024]
    end

    config.vm.define "devvm" do |devvm|
        devvm.vm.hostname = 'devvm'
        devvm.vm.network :private_network, ip: "10.0.123.2"
    end
end
