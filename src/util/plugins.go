package util

import (
	"errors"
	"net/http"
	"io"
	"fmt"
)


var GOTMLogger io.Writer
func InitPlugins() {
	GOTMLogger = StdoutFLog.Writer_m("GOTM.Log")
}

// may use r.URL.User in chat
// side effect executes request.ParseForm
func log(w HttpWriter, r HttpReq, info map[string]any) (render bool, addinfo any) {
	r.ParseForm()
	fmt.Fprintf(GOTMLogger, "new HTTP conn [%s] @ %s {%+v}\ninfo: %+v\n\n", r.Method, r.URL.Path, r.Form, info)
	return true, nil
}
var GOTM_log = GOTMPlugin{"", log}

func acc(w HttpWriter, r HttpReq, info map[string]any) (render bool, ret_r any) {
	ret_r = make(map[string]any)
	ret := ret_r.(map[string]any)
	render = true
	ret["ok"] = false

	ReadCookie, err := r.Cookie(AccountCookieName)
	if (errors.Is(err, http.ErrNoCookie)) {
		ret["cookieFail"] = true
		return
	}
	email, hash, ok := ReadAccountCookie(ReadCookie.Value)
	if (!ok) {
		ret["cookieFail"] = true
		ret["cookieSyntaxFail"] = true
		return
	}

	acc, exists := GetAccount(email)

	if (!exists) {
		ret["existsFail"] = true
		return
	}

	if (acc.Hash != hash) {
		ret["hashFail"] = true
		return
	}

	ret["name"] = acc.Name
	ret["email"] = acc.Email
	ret["ok"] = true
	return
}
var GOTM_account = GOTMPlugin{"acc", acc}

func url_to_info(w HttpWriter, r HttpReq, info map[string]any) (render bool, ret_r any) {
	ret_r = make(map[string]any)
	ret := ret_r.(map[string]any)

	//ret["scheme"] = r.URL.Scheme
	//ret["opaque"] = r.URL.Opaque
	ret["user"] = r.URL.User
	//ret["host"] = r.URL.Host
	ret["path"] = r.URL.Path
	//ret["rawPath"] = r.URL.RawPath
	//ret["forceQuery"] = r.URL.ForceQuery
	//ret["rawQuery"] = r.URL.RawQuery
	//ret["fragment"] = r.URL.Fragment
	//ret["rawFragment"] = r.URL.RawFragment
	//ret["string"] = r.URL.String()
	ret["query"] = r.URL.Query()
	//ret["unescapeQuery"], _ = url.QueryUnescape(r.URL.RawQuery)

	return true, ret_r
}
var GOTM_urlInfo = GOTMPlugin{"urlinfo", url_to_info}

func accountsCopy(w HttpWriter, r HttpReq, info map[string]any) (render bool, ret_r any) {
	return true, AccountsCopy()
}
var GOTM_accounts = GOTMPlugin{"accounts", accountsCopy}

// if a plugin runs an expensive funciton, that is only used if the used if
// logged in passes, it should check if hasacc == true before wasting time
// computing info that will be not rendered
func must_account(w HttpWriter, r HttpReq, info map[string]any) (render bool, ret_r any) {
	// if acc.ok == false ; /login
	if (!info["acc"].(map[string]any)["ok"].(bool)) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false, false
	}
	return true, true
}
var GOTM_mustacc = GOTMPlugin{"hasacc", must_account}

