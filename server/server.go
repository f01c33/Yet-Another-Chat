package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"./protocol"
)

type pkg struct {
	sender net.UDPAddr
	size   int
	data   []byte
}

type student struct {
	Nome    string
	Curso   string
	Senha   string
	Periodo int
	Addr    net.UDPAddr
	Sala    int
	tick    int
	votos   map[int]string
}

var (
	logger = log.New(os.Stdout, "[svr] ", log.LstdFlags)

	// blockSize = flag.Int("size", 1024, "block size to read packets on")
	blockSize = 2048

	service string

	db struct {
		Usuarios map[string]student
		mutex    *sync.Mutex
	}

	uList struct {
		Usuarios map[string]student
		mutex    *sync.Mutex
	}

	dList struct {
		Discussoes map[int]protocol.Discussao
		Votos      map[int]protocol.VotoStatus
		Mensagens  map[int][]protocol.Msg
		mutex      *sync.Mutex
	}
)

func str_to_hash(a string) string {
	hash := sha256.Sum256([]byte(a))

	bhash := []byte("")
	for _, v := range hash {
		// logger.Printf("%c:%c = %+v,%+v", i, v, i, v)
		bhash = append(bhash, v)
	}
	return string(hex.EncodeToString(bhash))
}

func load_global_vars() {
	uList.Usuarios = make(map[string]student, 30)
	uList.mutex = &sync.Mutex{}

	dList.Discussoes = make(map[int]protocol.Discussao, 0)
	dList.Mensagens = make(map[int][]protocol.Msg, 0)
	dList.Votos = make(map[int]protocol.VotoStatus, 0)
	dList.mutex = &sync.Mutex{}

	db.Usuarios = make(map[string]student, 30)
	db.mutex = &sync.Mutex{}
}

func program_data() {
	db.mutex.Lock() // fake students
	// uList.Usuarios["localhost"] = student{addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8888}}
	// uList.Usuarios["123"] = student{senha: "123", sala: -2, nome: "the_game"}
	db.Usuarios["123"] = student{Nome: "test", Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1687638"] = student{Nome: "Cauê Felchar", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1687824"] = student{Nome: "Vithor Tozetto Ferreira", Curso: "Ciência da Computação", Periodo: 7, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1600001"] = student{Nome: "Felipe Soares", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1687816"] = student{Nome: "Víctor Muniz dos Santos", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1488635"] = student{Nome: "Rodrigo", Curso: "Ciência da Computação", Periodo: 6, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1371886"] = student{Nome: "Gabriel Levis", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1591860"] = student{Nome: "Rafael Koteski", Curso: "Ciência da Computação", Periodo: 7, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1590529"] = student{Nome: "William Takeshi Omoto", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1592130"] = student{Nome: "Bruno Chagas", Curso: "Ciência da Computação", Periodo: 7, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1695738"] = student{Nome: "Diego Augusto Caldeira", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1687794"] = student{Nome: "Tiago Sanches Franco", Curso: "Ciência da Computação", Periodo: 8, Senha: str_to_hash("123"), Sala: -2}
	db.Usuarios["1547984"] = student{Nome: "Fernando Favero", Curso: "Ciência da Computação", Periodo: 5, Senha: str_to_hash("123"), Sala: -2}
	db.mutex.Unlock()

	file, err := os.Open("ipsum.txt")
	if err != nil {
		logger.Println(err)
	}
	defer file.Close()
	data := make([]byte, 100)
	_, err = file.Read(data)
	if err != nil {
		logger.Print(err)
	}

	arr := strings.Split(string(data), "\n")

	// logger.Print(arr)

	rand_str := func() string {
		l := rand.Intn(30)
		out := make([]byte, l)
		for i := 0; i < l; i++ {
			out[i] = byte(rand.Intn(90-65)) + 'a'
		}
		return string(out)
	}

	dList.mutex.Lock() // fake Discussoes
	for i := 0; i < len(arr); i++ {
		dList.Discussoes[i] = protocol.Discussao{
			Tipo:      4,
			ID:        i,
			Nome:      arr[i],
			Descricao: rand_str(),
			Criador:   rand_str(),
			Inicio:    strconv.FormatInt(time.Now().Add(time.Second*60).Unix(), 10),
			Fim:       strconv.FormatInt(time.Now().Add(time.Second*60).Unix(), 10),
			Status:    true,
		}
		dList.Mensagens[i] = make([]protocol.Msg, 0)
		dList.Votos[i] = protocol.VotoStatus{Tipo: 9, Acabou: false, Resultados: make([]map[string]int, 0)}
		vtmp := dList.Votos[i]
		for k := 0; k < rand.Intn(4); k++ {
			vtmp.Resultados = append(dList.Votos[i].Resultados, make(map[string]int, 0))
			vtmp.Resultados[len(vtmp.Resultados)-1][rand_str()] = 0
		}
		dList.Votos[i] = vtmp
		for j := 0; j < 10; j++ {
			dList.Mensagens[i] = append(dList.Mensagens[i], protocol.Msg{
				Tipo:      12,
				ID:        j,
				Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
				Criador:   rand_str(),
				Mensagem:  rand_str(),
			})
		}
		// logger.Print()
	}
	dList.mutex.Unlock()
}

func send_json(a net.UDPAddr, ch chan pkg, j interface{}) {
	otp, err := json.Marshal(j)
	if err != nil {
		logger.Println("erro:", err)
	}
	ch <- pkg{a, len(otp), otp}
}

func main() {
	/* to-do:
	imprimir dinheiro
	*/
	load_global_vars()
	program_data()

	if len(os.Args) != 3 {
		logger.Fatalf("Usage: ./server host:port guiport")
	}

	service = os.Args[1]

	//ui stuff
	tmpl_main := template.Must(template.ParseFiles("srv.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl_main.Execute(w, nil)
	})

	tmpl := template.Must(template.ParseFiles("loop.html"))
	http.HandleFunc("/ajax", func(w http.ResponseWriter, r *http.Request) {
		var stuff struct {
			Usuarios   map[string]student
			Discussoes []protocol.Discussao
		}
		uList.mutex.Lock()
		dList.mutex.Lock()
		discs := make([]protocol.Discussao, len(dList.Discussoes))
		for k, v := range dList.Discussoes {
			v.Tamanho = len(dList.Mensagens[k])

			tmp, err := strconv.ParseInt(v.Inicio, 10, 64)
			if err != nil {
				logger.Print(err)
			}
			v.Inicio = time.Unix(tmp, 0).Format(time.RFC822)

			tmp, err = strconv.ParseInt(v.Fim, 10, 64)
			if err != nil {
				logger.Print(err)
			}
			v.Fim = time.Unix(tmp, 0).Format(time.RFC822)
			discs[k] = v
			// logger.Print(v)
		}

		stuff.Usuarios = uList.Usuarios
		stuff.Discussoes = discs
		tmpl.Execute(w, stuff)
		dList.mutex.Unlock()
		uList.mutex.Unlock()
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	go http.ListenAndServe(":"+os.Args[2], nil)
	//end ui stuff

	// init
	flag.Parse()

	// to recieve udp packages from co-routine
	input := make(chan pkg, 10000)
	output := make(chan pkg, 10000)
	// for checking if Usuarios are alive

	// ip := net.ParseIP(*host)
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// thread to listen to udp packages
	go func(conn *net.UDPConn, in chan pkg) {
		logger.Printf("listening on addr=%s with block size=%d", conn.LocalAddr(), blockSize)
		data := make([]byte, blockSize)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				logger.Fatalf("error during read: %s", err)
			}
			logger.Printf("in:%s", data[:n])
			in <- pkg{*remoteAddr, n, data}
		}
	}(conn, input)

	// thread to send udp packages
	go func(conn *net.UDPConn, out chan pkg) {
		for {
			select {
			case p := <-out:
				logger.Printf("out:%s", p.data)
				_, err := conn.WriteToUDP(p.data, &p.sender)
				if err != nil {
					logger.Print("out error: ", err)
				}
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}(conn, output)

	erro := func(p pkg, o chan pkg) {
		send_json(p.sender, o, protocol.Erropkg{Tipo: -1, Pacote: string(p.data[:p.size])})
	}
	// var broadcast bool = false

	ticker := time.NewTicker(time.Second * 13).C

	var tipo protocol.Base
	for {
		select {
		case p := <-input:
			// logger.Print("next!: ", string(p.data[:p.size]))
			err := json.Unmarshal(p.data[:p.size], &tipo)
			if err != nil {
				logger.Println("erro:", err)
				// continue
			}
			// fmt.Printf("%v", json.Valid(p.data[:p.size]))
			if json.Valid(p.data[:p.size]) {
				// logger.Print("invalid json")
				// erro(p, output)
				// send_json(p.sender, output, protocol.Erropkg{Tipo: -1, Pacote: string(p.data[:p.size])})
				// output
				// return error package
			}
			switch tipo.Tipo {
			case -1:
				logger.Print("erro!: ", p.data[:p.size])
			case 0: //login
				var login protocol.Login
				err := json.Unmarshal(p.data[:p.size], &login)
				if err != nil {
					logger.Println("erro:", err)
				}
				//checar senha/ra
				succ := false
				db.mutex.Lock()
				usr := db.Usuarios[login.Ra]
				db.mutex.Unlock()
				var nome string
				if usr.Senha == login.Senha {
					nome = usr.Nome
					usr.Sala = -1
					succ = true
					usr.Addr = p.sender
					usr.tick = 4
					usr.votos = make(map[int]string, 0)
				}
				uList.mutex.Lock()
				uList.Usuarios[login.Ra] = usr
				uList.mutex.Unlock()
				if succ { //certo
					dList.mutex.Lock()
					returnCode := protocol.Menu{Tipo: 2, Nome: nome, Tamanho: len(dList.Discussoes)}
					send_json(p.sender, output, returnCode)

					for k, v := range dList.Discussoes {
						v.Tamanho = len(dList.Mensagens[k])
						send_json(p.sender, output, v)
					}
					dList.mutex.Unlock()
				} else { //errado
					tipo.Tipo = 1
					send_json(p.sender, output, tipo)
				}
				break
			case 6: // cria sala
				var (
					sala protocol.NovaSala
					voto protocol.VotoStatus
					disc protocol.Discussao
				)
				err := json.Unmarshal(p.data[:p.size], &sala)
				if err != nil {
					logger.Print(err)
				}
				voto.Resultados = make([]map[string]int, 0)
				for _, v := range sala.Opcoes {
					voto.Resultados = append(voto.Resultados, make(map[string]int, 0))
					voto.Resultados[len(voto.Resultados)-1][v.Nome] = 0
				}
				disc.Tipo = 4

				uList.mutex.Lock()
				dList.mutex.Lock()
				disc.ID = len(dList.Discussoes)
				disc.Nome = sala.Nome
				disc.Descricao = sala.Descricao
				disc.Criador = "" //sala.Criador
				for _, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) {
						disc.Criador = v.Nome
					}
				}
				disc.Fim = sala.Fim
				disc.Inicio = strconv.FormatInt(time.Now().Unix(), 10)
				disc.Status = true //true=aberto false=fechado
				disc.Tamanho = 0
				dList.Discussoes[disc.ID] = disc
				dList.Votos[disc.ID] = voto
				// dList.Mensagens[len(dList.Mensagens)-1] = make([]protocol.Msg, 0)
				// uList.mutex.Unlock()

				//disc.Tipo = 4

				// logger.Println("h")
				// enviar pra todo mundo que tem uma sala nova
				// uList.mutex.Lock()
				for _, v := range uList.Usuarios {
					// if v.sala == -1 {
					send_json(v.Addr, output, disc)
					// logger.Println("bcast")
					// }
				}
				dList.mutex.Unlock()
				uList.mutex.Unlock()

				break
			case 13: //pede mensagem especifica
				var sala protocol.MsgAsk
				err := json.Unmarshal(p.data[:p.size], &sala)
				if err != nil {
					logger.Print(err)
				}
				dList.mutex.Lock()
				logger.Print("ask:", sala.IDsala, sala.IDmsg)
				send_json(p.sender, output, dList.Mensagens[sala.IDsala][sala.IDmsg])
				dList.mutex.Unlock()
				break
			case 5: //pede sala especifica
				var sala protocol.SalaAsk
				err := json.Unmarshal(p.data[:p.size], &sala)
				if err != nil {
					logger.Print(err)
				}
				dList.mutex.Lock()
				send_json(p.sender, output, dList.Discussoes[sala.IDsala])
				dList.mutex.Unlock()
				break
			case 3: //logout
				uList.mutex.Lock()
				for k, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) {
						delete(uList.Usuarios, k)
						break
					}
				}
				uList.mutex.Unlock()
				break
			case 7:
				var pedido protocol.Acesso
				err := json.Unmarshal(p.data[:p.size], &pedido)
				if err != nil {
					logger.Print(err)
				}
				users := make([]protocol.StrObjHist, 0)
				uList.mutex.Lock()
				usr := protocol.Connect{Tipo: 10, Adicionar: true}
				for k, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) { // troca sala do usuário
						v.Sala = pedido.ID
						uList.Usuarios[k] = v
						usr.Nome = v.Nome
						usr.Ra = k
					} else if v.Sala == pedido.ID { // coloca quem está na sala na lista pra enviar, ignorando o proprio usuario
						users = append(users, protocol.StrObjHist{Nome: v.Nome, RA: k})
					}

				}
				uList.mutex.Unlock()
				dList.mutex.Lock()
				//historico
				send_json(p.sender, output, protocol.Historico{
					Tipo:     8,
					Tamanho:  len(dList.Mensagens[pedido.ID]),
					Usuarios: users,
				})
				for _, v := range dList.Mensagens[pedido.ID] {
					v.Tamanho = len(dList.Mensagens[pedido.ID])
					send_json(p.sender, output, v)
				}
				//voto
				vote := dList.Votos[pedido.ID]
				vote.Tipo = 9
				for k := range vote.Resultados {
					// vote.Resultados[k] = make(map[string]int, 0)
					for j := range vote.Resultados[k] {
						vote.Resultados[k][j] = 0
					}
				}
				send_json(p.sender, output, vote)

				dList.mutex.Unlock()

				uList.mutex.Lock()
				for _, v := range uList.Usuarios {
					if v.Sala == pedido.ID {
						send_json(v.Addr, output, usr)
					}
				}
				uList.mutex.Unlock()
				break
			case 14:
				var newmsg protocol.Msg
				err := json.Unmarshal(p.data[:p.size], &newmsg)
				if err != nil {
					logger.Print(err)
				}
				dList.mutex.Lock()
				uList.mutex.Lock()
				room := -2
				for _, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) {
						room = v.Sala
						newmsg.Criador = v.Nome
					}
				}
				if room == -2 {
					erro(p, output)
					break
				}
				if dList.Discussoes[room].Status == false {
					send_json(p.sender, output, dList.Votos[room])
					uList.mutex.Unlock()
					dList.mutex.Unlock()
					break
				}
				// logger.Print("room: ", room, "sender: ", p.sender, "usrs: ", uList.Usuarios, "\n")
				newmsg.Tipo = 12
				newmsg.ID = len(dList.Mensagens[room])
				newmsg.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
				newmsg.Tamanho = newmsg.ID + 1
				dList.Mensagens[room] = append(dList.Mensagens[room], newmsg)
				//bcast msg
				for _, v := range uList.Usuarios {
					send_json(v.Addr, output, newmsg)
				}
				uList.mutex.Unlock()
				dList.mutex.Unlock()
				// logger.Print("here!")
				break
			case 11:
				uList.mutex.Lock()
				usr := protocol.Connect{Tipo: 10, Adicionar: false}
				for k, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) {
						v.Sala = -2
						uList.Usuarios[k] = v
						usr.Nome = v.Nome
						usr.Ra = k
					}
				}
				for _, v := range uList.Usuarios {
					send_json(v.Addr, output, usr)
				}
				uList.mutex.Unlock()
				break
			case 16:
				var ping protocol.Ping
				json.Unmarshal(p.data[:p.size], &ping)
				uList.mutex.Lock()
				for k, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) {
						// usr := uList.Usuarios[k]
						// usr.tick = 4
						v.tick = 4
						uList.Usuarios[k] = v
						break
					}
				}
				uList.mutex.Unlock()
				break
			case 15:
				var vote protocol.Voto
				err := json.Unmarshal(p.data[:p.size], &vote)
				if err != nil {
					logger.Print(err)
				}
				dList.mutex.Lock()
				if dList.Discussoes[vote.Sala].Status == false {
					send_json(p.sender, output, dList.Votos[vote.Sala])
					dList.mutex.Unlock()
					break
				}
				dList.mutex.Unlock()
				uList.mutex.Lock()
				for k, v := range uList.Usuarios {
					if v.Addr.IP.Equal(p.sender.IP) {
						v.votos[vote.Sala] = vote.Opcao
						uList.Usuarios[k] = v
						send_json(p.sender, output, vote)
						break
					}
				}
				uList.mutex.Unlock()
				break
			default:
				erro(p, output)
			}
			// output <- pkg{p.sender, len(strconv.Itoa(tipo.Tipo)), []byte(strconv.Itoa(tipo.Tipo))}
		case <-ticker:
			// logger.Print("tick")
			uList.mutex.Lock()
			for k := range uList.Usuarios {
				if uList.Usuarios[k].tick == 1 { // enviar pra sala dele que ele saiu!
					for l := range uList.Usuarios {
						if l != k {
							send_json(uList.Usuarios[l].Addr, output, protocol.Connect{
								Tipo:      10,
								Adicionar: false,
								Nome:      uList.Usuarios[k].Nome,
								Ra:        k,
							})
						}
					}
					delete(uList.Usuarios, k)
				}
				usr := uList.Usuarios[k]
				usr.tick--
				uList.Usuarios[k] = usr
			}
			uList.mutex.Unlock()
		default:
			//ticking away, the moments that make up a dull day
			dList.mutex.Lock()
			for k, v := range dList.Discussoes {
				tmp, err := strconv.ParseInt(v.Fim, 10, 64)
				if err != nil {
					logger.Print(err)
				}
				if time.Until(time.Unix(tmp, 0)) <= 0 && v.Status == true {
					v.Status = false
					vt := dList.Votos[v.ID]
					vt.Acabou = true
					uList.mutex.Lock()
					for _, m := range uList.Usuarios {
						for n := range vt.Resultados {
							for o := range vt.Resultados[n] {
								if o == m.votos[v.ID] {
									vt.Resultados[n][o]++
									// vt.Resultados[m.votos[v.ID]]++
								}
							}
						}
					}
					dList.Votos[v.ID] = vt
					for _, m := range uList.Usuarios {
						if m.Sala == v.ID {
							logger.Print("here0")
							send_json(m.Addr, output, vt)
						}
					}
					uList.mutex.Unlock()
					dList.Discussoes[k] = v
				}
			}
			dList.mutex.Unlock()
			// time.Sleep(100 * time.Millisecond)
		}
	}
}
