package pool

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

//var PoolCreateExsample = InitPool(1, 1, 10, func() interface{} {
//	return &http.Request{} //any http client you want use
//}, time.Minute)

type HttpConnPool struct {
	m         *sync.RWMutex
	recycling time.Duration
	max,
	idleMax,
	inuse,
	free int
	insfunc insFun
	pool    []*clientConn
}

type insFun func() interface{}
type clientConn struct {
	Agent  interface{}
	enable bool
}

func (p *HttpConnPool) String() string {
	return fmt.Sprintf("free:%v, inuse:%v  recycling duration:%v", p.free, p.inuse, p.recycling)
}
func (p *HttpConnPool) Countreal() int {
	p.m.RLock()
	defer p.m.RUnlock()
	return len(p.pool)

}
func InitPool(l, i, max int, fun insFun, du time.Duration) *HttpConnPool {
	p := new(HttpConnPool)
	p.m = new(sync.RWMutex)
	p.m.Lock()
	defer p.m.Unlock()
	v := reflect.ValueOf(fun())
	if v.Kind() != reflect.Ptr {
		panic("the InsFun only accept the pointer of the client instance!!")
	}
	p.insfunc = fun
	p.recycling = du
	//fmt.Println("pool init ok")
	for ; l > 0; l-- {
		//fmt.Println("inti pool loop")
		p.pool = append(p.pool, &clientConn{p.insfunc(), true})
	}
	p.free = l
	p.max = max
	p.idleMax = i
	go func() {
		for {
			p.m.Lock()

			if p.free > p.idleMax {
				for index, conn := range p.pool {
					if conn.enable {
						p.remove(index)
						break
					}
				}

			}
			p.m.Unlock()
			t := time.NewTimer(du)
			<-t.C
		}
	}()
	fmt.Println("pool init ok")
	return p
}

func (p *HttpConnPool) remove(i int) {
	p.pool = append(p.pool[:i], p.pool[i+1:]...)
	p.free--
}

func (p *HttpConnPool) Get() (conn *clientConn, err error) {
	p.m.Lock()
	defer p.m.Unlock()
	if p.free > 0 {
		for _, conn := range p.pool {
			if conn.enable {
				{
					p.inuse++
					p.free--
				}
				conn.enable = false
				return conn, nil
			}
		}
	}

	if p.inuse < p.max {
		conn = &clientConn{p.insfunc(), false}
		p.pool = append(p.pool, conn)
		p.inuse++
		return
	} else {
		return nil, fmt.Errorf("pool is drain p.Free:%v  p.Inuse%v", p.free, p.inuse)
	}

}

func (p *HttpConnPool) Ret(conn *clientConn) {
	p.m.Lock()
	defer p.m.Unlock()
	conn.enable = true
	p.free++
	p.inuse--

}
func (c *clientConn) Ret(P *HttpConnPool) {
	P.m.Lock()
	defer P.m.Unlock()
	c.enable = true
	P.free++
	P.inuse--

}
