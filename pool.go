package pool

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

type HttpConnPool struct {
	m *sync.RWMutex
	recycling time.Duration
	max,
	idleMax,
	inuse,
	free int
	insfunc insFun
	pool    []*HttpConn
}

type insFun func() interface{}
type HttpConn struct {
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
func (p *HttpConnPool) InitPool(l, i, max int, fun insFun, du time.Duration) {
	p.m = new(sync.RWMutex)
	p.m.Lock()
	defer p.m.Unlock()
	v := reflect.ValueOf(fun())
	if v.Kind() != reflect.Ptr {
		panic("the InsFun only accept the pointer of the client instance!!")
	}
	p.insfunc = fun
	p.recycling=du
	//fmt.Println("pool init ok")
	for ; l > 0; l-- {
		fmt.Println("inti pool loop")
		p.pool = append(p.pool, &HttpConn{p.insfunc(), true})
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
}

func (p *HttpConnPool) remove(i int) {
	p.pool = append(p.pool[:i], p.pool[i+1:]...)
	p.free--
}

func (p *HttpConnPool) Get() (conn *HttpConn, err error) {
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
		conn = &HttpConn{p.insfunc(), false}
		p.pool = append(p.pool, conn)
		p.inuse++
		return
	} else {
		return nil, fmt.Errorf("pool is drain p.Free:%v  p.Inuse%v", p.free, p.inuse)
	}

}

func (p *HttpConnPool) Ret(conn *HttpConn) {
	p.m.Lock()
	defer p.m.Unlock()
	conn.enable = true
	p.free++
	p.inuse--

}
func (c *HttpConn) Ret(P *HttpConnPool) {
	P.m.Lock()
	defer P.m.Unlock()
	c.enable = true
	P.free++
	P.inuse--

}
