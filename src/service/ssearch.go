package service

import (
	"mysrv/util"
	"fmt"
	"strconv"
)

//var ServerFront = util.TemplatePage(
//	"html/ssearch.gohtml", nil,
//	[]util.GOTMPlugin{util.GOTM_account, util.GOTM_mustacc},
//)
//
//var ServerHTMX = util.LogicPage(
//	"", nil,
//	[]util.GOTMPlugin{util.GOTM_account, util.GOTM_mustacc},
//	htmxHandler,
//)

type option struct {
	DBid map[string]int // internal db to id mapping
	DBName string
	HTML func(string) string
	Name string // Name o HTML's input
	Options map[string]*option
}

func MakeSelection(
	formname string,
	labeltext string,
	dbname string,
	names ...any, // [(htmlname, textname)]
) *option {
	const colcount = 4
	if (len(names)%colcount!=0) {return nil}

	htmlnames := make([]string, len(names)/colcount)
	textnames := make([]string, len(names)/colcount)
	dbnames := make(map[string]int)
	optnames := make(map[string]*option)
	for index, name := range names {
		if (index%colcount == 0) {
			htmlnames[index/colcount] = name.(string)
		} else if (index%colcount == 1) {
			textnames[index/colcount] = name.(string)
		} else if (index%colcount == 2) {
			dbnames[htmlnames[(index-2)/colcount]] = name.(int)
		} else if (index%colcount == 3) {
			if (name != nil) {
				optnames[htmlnames[(index-3)/colcount]] = name.(*option)
			}
		}
	}
	HTMLFMT := fmt.Sprintf(`
	<div>
	<label for=%q>%s</label>
	<select id=%q name=%q>`, formname, labeltext, formname, formname)
	HTMLFMT+=`<option class="invisible" %s></option>`
	for index := range htmlnames {
		HTMLFMT+=fmt.Sprintf(
			`<option %%s value=%q> %s </option>`,
			htmlnames[index], textnames[index],
		)
	}
	HTMLFMT+="</select></div>"

	renderHTML := func (selected string) string {
		selection := make([]any, len(htmlnames)+1)
		selection[0] = bog(selected == "", "selected", "")
		for index, htmlname := range htmlnames {
			selection[index+1] = bog(selected == htmlname, "selected", "")
		}
		return fmt.Sprintf(HTMLFMT, selection...)
	}

	return &option{ dbnames, dbname, renderHTML, formname, optnames }
}

var opt_submit = &option{
	nil, "", func(s string) string {
		return `
<form
	id="sendinfo"
	hx-include="#info input, #info select"
	hx-post="/ssearch/htmx"
>
<button>Enviar</button>
</form>
`
	}, "", nil,
}

func bog(expr bool, yes, no string) string {
	if (expr) {
		return yes
	}
	return no
}

var opt_IaaS_Cloud_mem = MakeSelection(
	"IC-mem", "GigaBytes de RAM:",
	"Cloud_Server_Memory",
	"1",   "1 GB",   0, opt_submit,
	"2",   "2 GB",   1, opt_submit,
	"4",   "4 GB",   2, opt_submit,
	"8",   "8 GB",   3, opt_submit,
	"16",  "16 GB",  4, opt_submit,
	"32",  "32 GB",  5, opt_submit,
	"64",  "64 GB",  6, opt_submit,
	"128", "128 GB", 7, opt_submit,
)

var opt_IaaS_Cloud_cpu = MakeSelection(
	"IC-cpu", "quantidade de vCPUs:",
	"Cloud_Server_CPU",
	"1",   "1 vCPU",    0, opt_IaaS_Cloud_mem,
	"2",   "2 vCPUs",   1, opt_IaaS_Cloud_mem,
	"4",   "4 vCPUs",   2, opt_IaaS_Cloud_mem,
	"8",   "8 vCPUs",   3, opt_IaaS_Cloud_mem,
	"16",  "16 vCPUs",  4, opt_IaaS_Cloud_mem,
	"32",  "32 vCPUs",  5, opt_IaaS_Cloud_mem,
	"64",  "64 vCPUs",  6, opt_IaaS_Cloud_mem,
	"128", "128 vCPUs", 7, opt_IaaS_Cloud_mem,
)

var opt_IaaS_Cloud = MakeSelection(
	"IC-vm", "Sistema Operacional:",
	"Cloud_Server_VM",
	"linux", "VM Linux",             0, opt_IaaS_Cloud_cpu,
	"rhel", "VM Linux Corporativo", 1, opt_IaaS_Cloud_cpu,
	"windows", "VM Windows",           2, opt_IaaS_Cloud_cpu,
)

var opt_IaaS_Storage_Size = &option{
	nil, "Storage_Size",
	func (vl string) string {
		return fmt.Sprintf(`
		<input min="40" max="32768" name="ISS" type="number" value="%s">GB
		`, vl)
	}, "ISS",
	map[string]*option{"-": opt_submit},
}

var opt_IaaS_Storage = MakeSelection(
	"IS", "Serviço de Armazenamento:",
	"Storage_Type",
	"SSD", "Serviço de armazenamento de blocos (SSD)", 0, opt_IaaS_Storage_Size,
	"HDD", "Serviço de armazenamento de blocos (HDD)", 1, opt_IaaS_Storage_Size,
)

var opt_IaaS_Network_Trafic = MakeSelection(
	"INT", "Serviço de Tráfego",
	"Trafic_Type",
	"exit", "Tráfego de saída da rede",                   0, opt_submit,
	"between", "Tráfego de rede interna entre zonas",     1, opt_submit,
	"balance", "Tráfego de rede do balanceador de carga", 2, opt_submit,
)

var opt_IaaS_Network_Balance = MakeSelection(
	"INB", "Serviço de Balanceamento:",
	"Balance_Type",
	"load",     "Serviço de balanceamento de carga",                                                0, opt_submit,
	"DNS",      "Serviço de balanceamento de carga utilizando gerenciador de tráfego por DNS",      1, opt_submit,
	"endpoint", "Serviço de balanceamento de carga utilizando gerenciador de tráfego por endpoint", 2, opt_submit,
)

var opt_IaaS_Network_DNS = MakeSelection(
	"IND", "Serviço de DNS:",
	"DNS_Type",
	"zone",  "Serviço de DNS – Hospedagem de zonas", 0, opt_submit,
	"query", "Serviço de DNS – Consultas", 0, opt_submit,
)

var opt_IaaS_Network_Port = MakeSelection(
	"INP", "Serviço de Portas:",
	"Port_Speed",
	"1",  "Porta de conexão de fibra 1Gbps", 0, opt_submit,
	"10", "Porta de conexão de fibra 10Gbps", 0, opt_submit,
)
var opt_IaaS_Network = MakeSelection(
	"IN", "Serviço de rede:",
	"Network_Service_Type",
	"trafic", "Tráfego",        0, opt_IaaS_Network_Trafic,
	"balance", "Balanceamento", 1, opt_IaaS_Network_Balance,
	"DNS", "DNS",               2, opt_IaaS_Network_DNS,
	"port", "Porta",            3, opt_IaaS_Network_Port,
	"VPN", "VPN",               4, opt_submit,
	"IP", "IP Público",         5, opt_submit,
	"VDI", "VDI",               6, opt_submit,
)

var opt_IaaS_Security_Firewall = MakeSelection(
	"ISF", "Serviço de Firewall:",
	"Firewall_Type",
	"ACL", "Web Application Firewall por ACL",    1, opt_submit,
	"rule", "Web Application Firewall por Regra", 2, opt_submit,
	"hour", "Web Application Firewall por hora",  3, opt_submit,
)

var opt_IaaS_Security_Auth = MakeSelection (
	"ISA", "Serviço de Autenticação:",
	"Auth",
	"user", "Autenticação (Integração com AD) adquirido por usuário",   0, opt_submit,
	"domain", "Autenticação (Integração com AD) adquirido por domínio", 1, opt_submit,
)

var opt_IaaS_Security = MakeSelection(
	"IS", "Serviço de segurança:",
	"Firewall",
	"firewall", "Firewall",                     1, opt_IaaS_Security_Firewall,
	"backup",   "Backup",                       2, opt_submit,
	"storage",  "armazenamento de Backup",      3, opt_submit,
	"log",      "Auditoria de Análise de Logs", 4, opt_submit,
	"auth",     "Autenticação",                 5, opt_IaaS_Security_Auth,
	"password", "Cofre de senhas",              6, opt_submit,
)

var opt_IaaS = MakeSelection(
	"iaas-type", "Serviço de Infraestrutura:",
	"IaaS_Type",
	"cloud",    "Serviço de Computação em Nuvem", 0, opt_IaaS_Cloud,
	"storage",  "Armazenamento",                  1, opt_IaaS_Storage,
	"network",  "Rede",                           2, opt_IaaS_Network,
	"security", "Segurança",                      3, opt_IaaS_Security,
)

var opt_SaaS_Anal = MakeSelection(
	"SA", "Serviço de Analytics:",
	"Analitics_Type",
	"BI", "Serviço de BI (Visualização de Dados)", 0, opt_submit,
	"prot", "Serviço de proteção de dados utilizando endpoint por número de processador por servid", 1, opt_submit,
)

var opt_SaaS = MakeSelection(
	"SaaS-type", "Serviços de Software:",
	"Software_Type",
	"Anal", "Serviços de Analytics", 0, opt_SaaS_Anal,
	"CDN", "Serviço de distribuição de Conteúdo", 1, opt_submit,
)

var opt_PaaS_sql = MakeSelection(
	"PS", "Especificações do servidor:",
	"DB_Spec",
	"spec1",   "4 vCPUs e 16 GB de memória RAM", 0, opt_submit,
	"spec2",   "8 vCPUs e 32 GB de memória RAM", 1, opt_submit,
	"spec3",  "16 vCPUs e 64 GB de memória RAM", 2, opt_submit,
	"spec4", "32 vCPUs e 128 de GB memória RAM", 3, opt_submit,
)

var opt_PaaS_oracle = MakeSelection(
	"PS", "Especificações do servidor:",
	"DB_Spec",
	"spec1",   "4 vCPUs e 16 GB de memória RAM", 0, opt_submit,
	"spec2",   "8 vCPUs e 32 GB de memória RAM", 1, opt_submit,
	"spec3",  "16 vCPUs e 64 GB de memória RAM", 2, opt_submit,
)


var opt_PaaS = MakeSelection(
	"PaaS-type", "Serviços de Plataforma:",
	"DB_Type",
	"mysql", "MySQL", 0, opt_PaaS_sql,
	"postgresql", "PostgreSQL", 0, opt_PaaS_sql,
	"mssql", "SQLServer", 0, opt_PaaS_sql,
	"oracle", "ORACLE", 0, opt_PaaS_oracle,
)

var root = MakeSelection(
	"server-type", "Service Type:",
	"Service_Type",
	"IaaS", "IaaS", 0, opt_IaaS,
	"PaaS", "PaaS", 1, opt_PaaS,
	"SaaS", "SaaS", 2, opt_SaaS,
)

func htmxHandler(
	w util.HttpWriter, r util.HttpReq, info map[string]any,
) (render bool, addinfo any) {
	r.ParseForm()

	if (r.Method == "POST") {
		fmt.Fprintf(w, root.HTML(""))
		fmt.Fprintf(w, `<p style="color: green;">Request sent!</p>`)

		info := make(map[string]int)
		var node *option = root
		for node != opt_submit {
			v := r.FormValue(node.Name)
			info[node.DBName] = node.DBid[v]
			_node, ok := node.Options[v]
			if (!ok) {
				_node = node.Options["-"]
				intv, e := strconv.ParseInt(v, 10, 64)
				if (e != nil) {panic(e)}
				info[node.DBName] = int(intv)
			}
			node = _node
		}
		for k,v := range info {
			fmt.Printf("%q=>%d\n", k, v)
		}
		fmt.Println()
		return false, nil
	}

	var node *option = root
	for node != nil {
		v := r.FormValue(node.Name)
		fmt.Fprintf(w, node.HTML(v))
		_node, ok := node.Options[v]
		if (!ok && v != "") {
			_node = node.Options["-"]
		}
		node = _node
	}

	return false, nil // return render=false since there's no template to render
}

