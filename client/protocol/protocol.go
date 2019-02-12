package protocol

//Base for getting type
type Base struct {
	Tipo int `json:"tipo"`
}

// -1 = mensagem mal formada
//Erropkg -1 = Mensagem mal formada
type Erropkg struct {
	Tipo   int    `json:"tipo"`
	Pacote string `json:"pacote"`
	//	"Pacote":"Pacote inteiro erRAdo aqui pro caRA saber"
}

// 0 = login, do cliente pro servidor
//Login 0 = login, do cliente pro servIdor
type Login struct {
	Tipo  int    `json:"tipo"`
	Ra    string `json:"ra"`
	Senha string `json:"senha"`
}

// 1 = login, do servidor, dando errado
//base
// 2 = lista de discussões, do servidor, login bem sucedido
//Menu 2 = lista de discussões, do servIdor, login bem sucedIdo
type Menu struct {
	Tipo    int    `json:"tipo"`
	Nome    string `json:"nome"`
	Tamanho int    `json:"tamanho"`
}

// 3 = logout, enviado do cliente ao se desconectar
//base
// 4 = uma discussão
//Discussao 11 = uma discussão - Atual 4
type Discussao struct {
	Tipo      int    `json:"tipo"`
	ID        int    `json:"id"`
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Criador   string `json:"criador"`
	Inicio    string `json:"inicio"`
	Fim       string `json:"fim"`
	Status    bool   `json:"status"`
	Tamanho   int    `json:"tamanho"`
}

type Discs []Discussao

func (a Discs) Len() int           { return len(a) }
func (a Discs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Discs) Less(i, j int) bool { return a[i].ID < a[j].ID }

// 5 = pedir sala especifica
//SalaAsk 14 = pedir Sala especifica - Atual 5
type SalaAsk struct {
	Tipo   int `json:"tipo"`
	IDsala int `json:"id_sala"`
}

//Strobj pRA pegar Sala e outros
type StrObj struct {
	Nome string `json:"nome"`
}

//Strobj pRA pegar Sala e outros
type StrObjHist struct {
	Nome string `json:"nome"`
	RA   string `json:"ra"`
}

// 6 = criar sala
//NovaSala 3 = criar Sala - Atual 6
type NovaSala struct {
	Tipo int `json:"tipo"`
	// Criador   string   `json:"criador"`
	Nome      string   `json:"nome"`
	Descricao string   `json:"descricao"`
	Fim       string   `json:"fim"`
	Opcoes    []StrObj `json:"opcoes"`
}

// 7 = cliente pedindo acesso a sala
//Acesso 5 = cliente pedindo acesso a Sala - atual 7
type Acesso struct {
	Tipo int `json:"tipo"`
	ID   int `json:"id"`
}

// 8 = historico e usuários, do servidor
//Historico 6 = historico e usuários, do servIdor - Atual 8
type Historico struct {
	Tipo     int          `json:"tipo"`
	Tamanho  int          `json:"tamanho"`
	Usuarios []StrObjHist `json:"usuarios"`
}

// 9 = status da votação
//VotoStatus 7 = Status da votação - Atual 9
type VotoStatus struct {
	Tipo       int              `json:"tipo"`
	Acabou     bool             `json:"acabou"`
	Resultados []map[string]int `json:"resultados"`
}

// 10 = desconectar/conectar usuário
//Connect 16 = desconectar/conectar usuário - Atual 10
type Connect struct {
	Tipo      int    `json:"tipo"`
	Adicionar bool   `json:"adicionar"`
	Nome      string `json:"nome"`
	Ra        string `json:"ra"`
}

// 11 = logout da sala do cliente para
//base

// 12 = mensagem do servidor
//Msg 9 = Mensagem do servIdor - Indice 12
type Msg struct {
	Tipo      int    `json:"tipo"`
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Tamanho   int    `json:"tamanho"`
	Criador   string `json:"criador"`
	Mensagem  string `json:"mensagem"`
}
type Msgs []Msg

func (a Msgs) Len() int           { return len(a) }
func (a Msgs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Msgs) Less(i, j int) bool { return a[i].ID < a[j].ID }

// 13 = pedir mensagem especifica
//MsgAsk 13 = pedir Mensagem especifica
type MsgAsk struct {
	Tipo   int `json:"tipo"`
	IDmsg  int `json:"id_msg"`
	IDsala int `json:"id_sala"`
}

// 14 = mensagem do cliente pro servidor
//MsgInput 8 = Mensagem do cliente pro servIdor - atual 14
type MsgInput struct {
	Tipo     int    `json:"tipo"`
	Criador  string `json:"criador"`
	Mensagem string `json:"mensagem"`
}

// 15 = voto
//Voto 15 = voto
type Voto struct {
	Tipo  int    `json:"tipo"`
	Sala  int    `json:"sala"`
	Opcao string `json:"opcao"`
}

// 16 = ping
//Ping  -2 = ping - Atual 16
type Ping struct {
	Tipo int `json:"tipo"`
	Sala int `json:"sala"`
}
