## Um programa de chat para a aula de sistemas distribuidos
Autores:
[Cauê]("https://github.com/f01c33")
[Walter]("https://github.com/walterBSG")

Para utilizar dos programas, entre nas respectivas pastas, pelo terminal, e rode:
go build (client/server).go
cliente:
	./client host:porta_comum porta_para_gui_web
servidor:
	./server host:porta_comum porta_para_gui_web

host pode ser tanto localhost como pode ser o ip local, que pode ser encontrado com o comando:
	ifconfig | grep 'inet '

para instalação, instale go: https://golang.org/doc/install

para o manjaro/arch, um simples :
	sudo pacman -S go
é o bastante para instalar, em outros linux, para instalar direto do .tar.gz:
	wget https://golang.org/doc/install?download=go1.9.2.linux-amd64.tar.gz;
  tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz;
	cat 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile;