package neignbor

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
)

func isFoundHost(guessHost string,port uint16) bool{
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", guessHost, port))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

var Pattern = regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?\.){3})(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9])`)
func FindNeighbors(myHost string , myPort uint16, startIp int8,endIp uint8, startPort uint16,endPort uint16) []string{
	address:=fmt.Sprintf("%s:%d",myHost,myPort)
	m:=Pattern.FindStringSubmatch(myHost)
	if(m==nil){
		return nil
	}
	prefixHost :=m[1]
	lastIp,_:=strconv.Atoi(m[len(m)-1])
	neighbors:=make([]string,0)
	for port:=startPort; port<=uint16(endPort);port+=1{
		for ip:=startIp;ip<=int8(endIp);ip+=1{
			guessHost:=fmt.Sprintf("%s%d",prefixHost,lastIp+int(ip))
			guessTarget:=fmt.Sprintf("%s:%d",guessHost,port)
			if(guessTarget!=address && isFoundHost(guessHost,port)){
				neighbors=append(neighbors,guessTarget)
			}
		}
	}
	return neighbors
}

func GetHost() string{
	hostName,err:=os.Hostname()
	if(err!=nil){
		return "127.0.0.1"
	}
	address,err:=net.LookupHost(hostName)
	if err!=nil{
		return "127.0.0.1"
	}
	fmt.Println(address)
	return address[0]
}