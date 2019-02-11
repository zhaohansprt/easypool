package pool

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

//import "net/http"
//var PoolCreateExsample = InitPool(1, 1, 10, func(context.Context) (interface{},error) {
//	return &http.Request{},nil //any http client you want use
//}, time.Minute,context.Background())

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

type insFun func(context.Context) (interface{}, error)
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
func InitPool(l, i, max int, fun insFun, du time.Duration, ctx context.Context) (pool *HttpConnPool) {
	p := new(HttpConnPool)
	p.m = new(sync.RWMutex)
	p.m.Lock()
	defer p.m.Unlock()
	//必须执行一次才能校验出 用户传进来的fun 的返回值是不是符合规范
	v0, err0 := fun(ctx)
	if err0 != nil {
		panic(fmt.Sprintf("init pool error with initfun failed  error :", err0))
	}
	v := reflect.ValueOf(v0)
	if v.Kind() != reflect.Ptr {
		panic("the InsFun only accept  pointer type as the client instance!!")
	}
	p.insfunc = fun
	p.recycling = du
	//fmt.Println("pool init ok")
	for ; l > 0; l-- {
		//fmt.Println("inti pool loop")
		p.pool = append(p.pool, &clientConn{v0, true})
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

func (p *HttpConnPool) Get(ctx context.Context) (conn *clientConn, err error) {
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
		c, err := p.insfunc(ctx)
		if err != nil {
			return nil, err
		}
		conn = &clientConn{c, false}
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
