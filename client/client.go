package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"./protocol"
)

type Loginform struct {
	Ra      string
	Pwd     string
	Success bool
}

type pkg struct {
	sender net.UDPAddr
	size   int
	data   []byte
}

type states int

const (
	Start states = iota
	Login_wait
	Main
	NewDisc
	AtDisc
)

var (
	logger    = log.New(os.Stdout, "[clt] ", log.LstdFlags)
	blockSize = 2048
	server    = net.UDPAddr{}
	state     = Start
	nome      string
	sala      = -1
	conn      *net.UDPConn
	wait_cont = 0
	try_voto  = false
	voto      protocol.Voto

	uList struct {
		Usuarios []string
		mutex    *sync.Mutex
	}
	dList struct {
		Discussoes map[int]protocol.Discussao
		Votos      map[int]protocol.VotoStatus
		Mensagens  map[int][]protocol.Msg
		has_msg    map[int]map[int]bool
		mutex      *sync.Mutex
	}
)

func send_json(a net.UDPAddr, ch chan pkg, j interface{}) {
	otp, err := json.Marshal(j)
	if err != nil {
		logger.Println("erro:", err)
	}
	ch <- pkg{a, len(otp), otp}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port guiport\n", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]

	dList.mutex = &sync.Mutex{}
	uList.mutex = &sync.Mutex{}
	dList.Discussoes = make(map[int]protocol.Discussao, 0)
	dList.Mensagens = make(map[int][]protocol.Msg, 0)
	dList.has_msg = make(map[int]map[int]bool, 0)
	dList.Votos = make(map[int]protocol.VotoStatus)
	input := make(chan pkg, 10000)
	output := make(chan pkg, 10000)

	// loginData := protocol.Login{Ra: "", Senha: ""}
	// wait := template.Must(template., err)
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	tmpl := template.Must(template.ParseFiles("login.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch state {
		case Start:
			if r.Method != http.MethodPost {
				tmpl.Execute(w, nil)
				return
			}

			str_to_hash := func(a string) string {
				hash := sha256.Sum256([]byte(a))

				bhash := []byte("")
				for _, v := range hash {
					// logger.Printf("%c:%c = %+v,%+v", i, v, i, v)
					bhash = append(bhash, v)
				}
				return string(hex.EncodeToString(bhash))
			}

			loginData := protocol.Login{
				Ra:    r.FormValue("ra"),
				Senha: str_to_hash(r.FormValue("pwd")),
				// Senha: string(sha256.Sum256([]byte(r.FormValue("pwd")))),
				Tipo: 0,
			}

			send_json(server, output, loginData)

			state = Login_wait
			wait_cont = 0
			// tmpl.Execute(w, loginData)
			// time.Sleep(300 * time.Millisecond)
			http.Redirect(w, r, "/home", http.StatusFound)
		default:
			http.Redirect(w, r, "/home", http.StatusFound)
		}
	})

	tmplHome := template.Must(template.ParseFiles("home.html"))
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		switch state {
		case Start:
			http.Redirect(w, r, "/", http.StatusFound)
			// send packet
			// break
		case Login_wait:
			time.Sleep(time.Second)
			wait_cont++
			if wait_cont < 10 { //while not timeout
				http.Redirect(w, r, "/home", http.StatusFound)
			} else {
				state = Start
				http.Redirect(w, r, "/", http.StatusFound)
			}
			// tmplHome.Execute(w, nil)
		case Main:
			if r.Method != http.MethodPost {
				time.Sleep(50 * time.Millisecond)
				dList.mutex.Lock()

				mainmenu := struct {
					Nome       string
					Discussoes []protocol.Discussao
				}{
					Nome: nome,
				}

				mainmenu.Discussoes = make([]protocol.Discussao, len(dList.Discussoes))
				for k, v := range dList.Discussoes {
					mainmenu.Discussoes[k] = v
				}

				tmplHome.Execute(w, mainmenu)
				dList.mutex.Unlock()
			} else {
				state = NewDisc
				// logger
				// r.ParseMultipartForm(1024)
				r.ParseForm()
				logger.Print(r.Form)

				tmplHome.Execute(w, nil)

				//arrumar o botão de sair!

				// http.Redirect(w, r, "/nova", http.StatusFound)
			}
		}
	})

	tmplNova := template.Must(template.ParseFiles("nova.html"))
	http.HandleFunc("/nova", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmplNova.Execute(w, nil)
		} else {
			r.ParseForm()
			tostrobj := func(arr []string) []protocol.StrObj {
				out := make([]protocol.StrObj, len(arr))
				for k := range arr {
					out[k] = protocol.StrObj{Nome: arr[k]}
				}
				return out
			}
			fine := r.FormValue("fim")
			logger.Print("fim: ", fine)
			tmp := time.Now()
			if len(fine) > 0 {
				tmp, _ = time.Parse("2006-01-02T15:04", fine)
			}
			var newroom protocol.NovaSala = protocol.NovaSala{
				Tipo:      6,
				Nome:      r.FormValue("nome"),
				Descricao: r.FormValue("descricao"),
				Fim:       strconv.FormatInt(tmp.Unix(), 10),
				Opcoes:    tostrobj(strings.Split(r.FormValue("opcoes"), ",")),
			}
			send_json(server, output, newroom)
			state = Main
			http.Redirect(w, r, "/home", http.StatusFound)
		}
	})

	tmplogout := template.Must(template.ParseFiles("logout.html"))
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if state == Main {
			state = Start
			send_json(server, output, protocol.Base{3})
			tmplogout.Execute(w, nil)
		} else if state == AtDisc {
			state = Main
			send_json(server, output, protocol.Base{Tipo: 11})
			http.Redirect(w, r, "/home", http.StatusFound)
		}
	})

	tmplsala := template.Must(template.ParseFiles("disc.html"))
	http.HandleFunc("/disc/", func(w http.ResponseWriter, r *http.Request) {
		if state != Main && state != AtDisc {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		// r.
		room, err := strconv.Atoi(strings.Split(html.EscapeString(r.URL.Path), "/disc/")[1])
		sala = room
		if err != nil {
			logger.Print(err)
		}
		// protocol.Acesso{}
		if state == Main {
			send_json(server, output, protocol.Acesso{Tipo: 7, ID: room})
		}
		state = AtDisc
		// for len(dList.Votos[room].Resultados) == 0 {
		// 	time.Sleep(50 * time.Millisecond)
		// }
		dList.mutex.Lock()
		// uList.mutex.Lock()
		// iface := struct {
		// 	Disc  protocol.Discussao
		// 	Users []string
		// }{
		// 	Disc:  dList.Discussoes[room],
		// 	Users: []string{""},
		// }
		// if uList.Usuarios != nil {
		// 	iface.Users = uList.Usuarios
		// }
		// if iface.Disc == nil {
		// 	iface.Disc = protocol.Discussao{}
		// }
		// uList.mutex.Unlock()
		resp := struct {
			ID         int
			Nome       string
			Descricao  string
			Resultados map[string]int
			Votes      protocol.VotoStatus
		}{
			ID:        room,
			Nome:      dList.Discussoes[room].Nome,
			Descricao: dList.Discussoes[room].Descricao,
			Votes:     dList.Votos[room],
		}
		resp.Resultados = make(map[string]int, 0)
		for i := range dList.Votos[room].Resultados {
			for k, v := range dList.Votos[room].Resultados[i] {
				resp.Resultados[k] = v
			}
		}
		// logger.Print("Votes:", dList.Votos[room].Acabou)
		tmplsala.Execute(w, resp)
		dList.mutex.Unlock()
		// logger.Print(iface)

	})

	tmplmsgs := template.Must(template.ParseFiles("msgs.html"))
	http.HandleFunc("/ajax", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Print(err)
			return
		}
		dList.mutex.Lock()
		uList.mutex.Lock()
		room, _ := strconv.Atoi(string(data))

		resp := struct {
			Usrs []string
			Msgs []protocol.Msg
		}{
			Usrs: uList.Usuarios,
			Msgs: dList.Mensagens[room],
		}
		// logger.Printf("msgs:%+v\n", resp.Msgs[0].Criador)
		tmplmsgs.Execute(w, resp)
		// response := []byte("[")
		// for _, v := range uList.Usuarios {
		// 	response = append(response, []byte("{\"tipo\":8,\"nome\":"+v+"},")...)
		// }
		// for k, v := range dList.Mensagens[room] {
		// 	tmp, _ := json.Marshal(v)
		// 	response = append(response, tmp...)
		// 	if k < len(dList.Mensagens[room])-1 {
		// 		response = append(response, []byte(",")...)
		// 	}
		// }
		// response = append(response, []byte("]")...)
		// w.Write(response)
		uList.mutex.Unlock()
		dList.mutex.Unlock()
	})

	http.HandleFunc("/newmsg", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Print(err)
			return
		}
		send_json(server, output, protocol.MsgInput{Tipo: 14, Criador: nome, Mensagem: string(data)})
		// logger.Print("data!: ", data)
		// w.Write(data)
	})

	http.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Print(err)
			return
		}
		voto = protocol.Voto{Tipo: 15, Sala: sala, Opcao: string(data)}
		if try_voto == false {
			send_json(server, output, voto)
		}
		try_voto = true
	})

	http.HandleFunc("/end", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Print(err)
			return
		}
		room, _ := strconv.Atoi(string(data))
		dList.mutex.Lock()
		if dList.Votos[room].Acabou {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
		dList.mutex.Unlock()
	})

	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	go http.ListenAndServe(":"+os.Args[2], nil)

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	server = *udpAddr
	conn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// thread to listen to udp packages
	go func(conn *net.UDPConn, in chan pkg) {
		logger.Printf("listening on addr=%s with block size=%d", conn.LocalAddr(), blockSize)
		for {
			data := make([]byte, blockSize)
			n, remoteAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				logger.Print(n, remoteAddr, err)
				udpAddr, err := net.ResolveUDPAddr("udp4", service)
				if err != nil {
					logger.Fatalf("error during read: %s", err)
				}
				server = *udpAddr
				conn, err = net.DialUDP("udp", nil, udpAddr)
				continue
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
				// conn.WriteToUDP(p.data, &p.sender)
				conn.Write(p.data)
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}(conn, output)

	//pinger
	go func(o chan pkg) {
		// ticker := time.NewTicker(time.Second * 6).C
		for {
			if state != Start {
				// 	select {
				// 	case <-ticker:
				time.Sleep(time.Second * 12)
				send_json(server, o, protocol.Ping{Tipo: 16, Sala: sala})
				// default:
				// }
			}
		}
	}(output)

	var tipo protocol.Base
	var max struct {
		val int
		has bool
	}
	max.has = false

	erro := func(p pkg, o chan pkg) {
		logger.Print("Erro? ", string(p.data[:p.size]))
		// send_json(p.sender, o, protocol.Erropkg{Tipo: -1, Pacote: string(p.data[:p.size])})
	}
	cont_miss := 0
	for {
		// fmt.Printf("%v", json.Valid(p.data[:p.size]))
		// if json.Valid(p.data[:p.size]) {
		// output
		// return error package
		// }
		switch state {
		case Start:
			select {
			case p := <-input:
				err := json.Unmarshal(p.data[:p.size], &tipo)
				if err != nil {
					logger.Println("erro:", err)
				}
				if tipo.Tipo == 4 {
					var disc protocol.Discussao
					err := json.Unmarshal(p.data[:p.size], &disc)
					if err != nil {
						logger.Fatal("wait_3", err)
					}
					dList.mutex.Lock()
					dList.Discussoes[disc.ID] = disc
					// dList.Discussoes = append(dList.Discussoes, disc)
					// sort.Sort(protocol.Discs(dList.Discussoes))
					dList.mutex.Unlock()
				} else {
					erro(p, output)
				}
			default:
				time.Sleep(20 * time.Millisecond)
				// logger.Print("Start_default")
			}
		case Login_wait:
			time.Sleep(20 * time.Millisecond)
			select {
			case p := <-input:
				// logger.Print("watwat")
				err = json.Unmarshal(p.data[:p.size], &tipo)
				if err != nil {
					// fmt.Printf("'%v,%+v,%+v'\n", "wait_1", err, string(p.data))
				} else if tipo.Tipo == 1 {
					state = Start
					break
				} else if tipo.Tipo == 2 {
					var mainmenu protocol.Menu
					err = json.Unmarshal(p.data[:p.size], &mainmenu)
					if err != nil {
						logger.Print("wait_1_tipo2")
					}
					max.has = true
					max.val = mainmenu.Tamanho
					nome = mainmenu.Nome
				} else if tipo.Tipo == 4 {
					var disc protocol.Discussao
					err := json.Unmarshal(p.data[:p.size], &disc)
					if err != nil {
						// logger.Fatal("wait_3", err)
					}
					disc.Tipo = 4
					// logger.Print("here?", disc)
					dList.mutex.Lock()
					dList.Discussoes[disc.ID] = disc
					// dList.Discussoes[] = disc)
					// logger.Print("disc_here: ", dList.Discussoes)
					dList.mutex.Unlock()
				} else {
					erro(p, output)
					// logger.Print("wat: ", p.data[:p.size])
				}
				break
			default:
				dList.mutex.Lock()
				// logger.Print("wat?", max, len(dList.Discussoes))
				if max.has == true {
					if len(dList.Discussoes) == max.val {
						// sort.Sort(protocol.Discs(dList.Discussoes))
						state = Main
					} else {
						disc := make([]bool, len(dList.Discussoes))
						for i := 0; i < max.val; i++ {
							disc[i] = false
						}
						for _, v := range dList.Discussoes {
							// logger.Print("d: ", v)
							// disc = append(disc,v.ID)
							disc[v.ID] = true
						}
						for a := range disc {
							if disc[a] == false {
								send_json(server, output, protocol.SalaAsk{Tipo: 14, IDsala: a})
							}
						}
					}
				}
				dList.mutex.Unlock()
				time.Sleep(2 * time.Second)
			}
		case AtDisc:
			time.Sleep(100 * time.Millisecond)
			select {
			case p := <-input:
				// logger.Print("watwat")
				err = json.Unmarshal(p.data[:p.size], &tipo)
				if err != nil {
					logger.Print(err)
				}
				if tipo.Tipo == 12 {
					var msg protocol.Msg
					err := json.Unmarshal(p.data[:p.size], &msg)
					if err != nil {
						logger.Print("wait_3", err)
					}
					// logger.Print("msg:", msg.ID)
					dList.mutex.Lock()
					if dList.Mensagens[sala] == nil {
						dList.Mensagens[sala] = make([]protocol.Msg, 0)
						dList.has_msg[sala] = make(map[int]bool, 0)
					}
					if dList.has_msg[sala][msg.ID] == false {
						dList.Mensagens[sala] = append(dList.Mensagens[sala], msg)
						sort.Sort(protocol.Msgs(dList.Mensagens[sala]))
						dList.has_msg[sala][msg.ID] = true
					}
					dList.mutex.Unlock()
				} else if tipo.Tipo == 8 {
					var hist protocol.Historico
					err := json.Unmarshal(p.data[:p.size], &hist)
					if err != nil {
						logger.Print(err)
					}
					uList.mutex.Lock()
					uList.Usuarios = make([]string, 0)
					for _, v := range hist.Usuarios {
						logger.Print("adding: ", v.Nome)
						uList.Usuarios = append(uList.Usuarios, v.Nome)
					}
					uList.Usuarios = append(uList.Usuarios, nome)
					uList.mutex.Unlock()
				} else if tipo.Tipo == 9 {
					var vote protocol.VotoStatus
					err := json.Unmarshal(p.data[:p.size], &vote)
					if err != nil {
						logger.Print(err)
					}
					dList.mutex.Lock()
					dList.Votos[sala] = vote
					dList.mutex.Unlock()
				} else if tipo.Tipo == 10 {
					var newcon protocol.Connect
					err = json.Unmarshal(p.data[:p.size], &newcon)
					if err != nil {
						logger.Print(err)
					}
					uList.mutex.Lock()
					if newcon.Adicionar {
						exist := false
						for _, v := range uList.Usuarios {
							if v == newcon.Nome {
								exist = true
							}
						}
						if exist == false {
							uList.Usuarios = append(uList.Usuarios, newcon.Nome)
						}
					} else {
						for k := range uList.Usuarios {
							if uList.Usuarios[k] == newcon.Nome {
								//remove user
								uList.Usuarios[k] = uList.Usuarios[len(uList.Usuarios)-1]
								uList.Usuarios = uList.Usuarios[:len(uList.Usuarios)-1]
								break
							}
						}
					}
					uList.mutex.Unlock()
				} else if tipo.Tipo == 15 {
					var vote protocol.Voto
					err = json.Unmarshal(p.data[:p.size], &vote)
					if err != nil {
						logger.Print(err)
					}
					if voto.Opcao == vote.Opcao {
						try_voto = false
					}
				} else {
					erro(p, output)
				}
			default:
				dList.mutex.Lock() // check for missing messages
				if len(dList.Mensagens[sala])+1 < dList.Discussoes[sala].Tamanho {
					// logger.Print(len(dList.Mensagens[sala]), dList.Discussoes[sala].Tamanho)
					for i := 0; i < dList.Discussoes[sala].Tamanho && i < len(dList.Mensagens[sala]) && cont_miss > dList.Discussoes[sala].Tamanho; i++ {
						// if dList.Mensagens[sala][i].ID != i {
						if dList.has_msg[sala][i] == false {
							// logger.Print(i, dList.Mensagens[sala][i].ID)
							send_json(server, output, protocol.MsgAsk{Tipo: 13, IDmsg: i, IDsala: sala})
						}
					}
					cont_miss++
				}
				if try_voto { //hasn't received server update on vote yet
					if cont_miss > 3 {
						send_json(server, output, voto)
						time.Sleep(time.Second)
					}
					cont_miss++
				}
				dList.mutex.Unlock()
				if cont_miss >= dList.Discussoes[sala].Tamanho {
					time.Sleep(time.Duration(cont_miss*1) * time.Second)
				}
				break
			}
			break
		default:
			time.Sleep(200 * time.Millisecond)
			// logger.Print("unexpected")
			select {
			case p := <-input:
				// logger.Print("watwat")
				err = json.Unmarshal(p.data[:p.size], &tipo)
				if err != nil {
					logger.Fatal(err)
				}
				if tipo.Tipo == 4 {
					var disc protocol.Discussao
					err := json.Unmarshal(p.data[:p.size], &disc)
					if err != nil {
						logger.Fatal("wait_3", err)
					}
					dList.mutex.Lock()
					dList.Discussoes[disc.ID] = disc
					// dList.Discussoes = append(dList.Discussoes, disc)
					// sort.Sort(protocol.Discs(dList.Discussoes))
					dList.mutex.Unlock()
				}
			}
		}
		// output <- pkg{p.sender, len(strconv.Itoa(tipo.Tipo)), []byte(strconv.Itoa(tipo.Tipo))}
	}
}
