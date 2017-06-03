package vm

import (
	"fmt"
	"github.com/fatih/structs"
	"io"
	"log"
	"net/http"
	"strings"
)

type response struct {
	status int
	body   string
}

type request struct {
	Method string
	Body   string
	URL    string
	Path   string
}

func initializeSimpleServerClass(vm *VM) {
	initializeHTTPClass(vm)
	net := vm.loadConstant("Net", true)
	simpleServer := initializeClass("SimpleServer", false)
	simpleServer.setBuiltInMethods(builtinSimpleServerClassMethods, true)
	simpleServer.setBuiltInMethods(builtinSimpleServerInstanceMethods, false)
	net.constants[simpleServer.Name] = &Pointer{simpleServer}
}

var builtinSimpleServerClassMethods = []*BuiltInMethodObject{
	{
		Name: "new",
		Fn: func(receiver Object) builtinMethodBody {
			return func(v *VM, args []Object, blockFrame *callFrame) Object {
				serverClass := v.constants["Net"].returnClass().constants["SimpleServer"].returnClass()
				server := serverClass.initializeInstance()
				server.InstanceVariables.set("@port", args[0])
				return server
			}
		},
	},
}

var builtinSimpleServerInstanceMethods = []*BuiltInMethodObject{
	{
		Name: "start",
		Fn: func(receiver Object) builtinMethodBody {
			return func(v *VM, args []Object, blockFrame *callFrame) Object {
				var port string

				portVar, ok := receiver.(*RObject).InstanceVariables.get("@port")

				if !ok {
					port = "8080"
				} else {
					port = portVar.(*StringObject).Value
				}

				fmt.Println("Start listening on port: " + port)
				log.Fatal(http.ListenAndServe(":"+port, nil))
				return receiver
			}
		},
	},
	{
		Name: "mount",
		Fn: func(receiver Object) builtinMethodBody {
			return func(v *VM, args []Object, blockFrame *callFrame) Object {
				path := args[0].(*StringObject).Value

				http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
					req := initRequest(r)
					res := httpResponseClass.initializeInstance()

					v.builtInMethodYield(blockFrame, req, res)

					setupResponse(w, r, res)
				})

				return receiver
			}
		},
	},
}

func initRequest(req *http.Request) *RObject {
	r := request{}
	reqObj := httpRequestClass.initializeInstance()

	r.Method = req.Method
	r.Body = ""
	r.Path = req.URL.Path
	r.URL = req.Host + req.RequestURI

	m := structs.Map(r)

	for k, v := range m {
		varName := "@" + strings.ToLower(k)
		reqObj.InstanceVariables.set(varName, initObject(v))
	}

	return reqObj
}

func setupResponse(w http.ResponseWriter, req *http.Request, res *RObject) {
	r := &response{}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header

	resStatus, ok := res.InstanceVariables.get("@status")

	if ok {
		r.status = resStatus.(*IntegerObject).Value
	} else {
		r.status = http.StatusOK
	}

	resBody, ok := res.InstanceVariables.get("@body")

	if !ok {
		r.body = ""
	} else {
		r.body = resBody.(*StringObject).Value
	}

	io.WriteString(w, r.body)
	fmt.Printf("%s %s %s %d\n", req.Method, req.URL.Path, req.Proto, r.status)
}

func initObject(v interface{}) Object {
	switch v := v.(type) {
	case string:
		return initializeString(v)
	case int:
		return initilaizeInteger(v)
	case bool:
		if v {
			return TRUE
		}

		return FALSE
	default:
		panic("Can't init object")
	}
}
