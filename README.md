# easypool


## features:

**1.并发安全  concurrent safe**

**2.自动回收  auto recycle the free clients**



## 用例

**初始化  初始client数量 ， 最小闲置数量 ，最大数量 ，client 实例创建的回调函数， 回收间隔**

```
ctx := context.Background()
ReqHttpConnP,error=pool.InitPool(1,80,300,  func(context.Context) (interface{},error) {return gorequest.New(),nil},time.Second*10, ctx)  
```

**获取**
```
conn,err:=ReqHttpConnP.Get(ctx)
```

**归还**
```
conn.Ret(ReqHttpConnP)
```





**取出以后需要强制转换成你放进去的类型**


```
func  Test_sendreward(t *testing.T)  {

	ret:=new(RspPhp)
	fun:=func (resp gorequest.Response, body []byte, errs []error){
		if resp.StatusCode==200{
			if err:=json.Unmarshal(body,ret);err!=nil{
				fmt.Printf("sendreward to php decode rsp failed error:%v  retmsg:%v  \n",err,  string(body))
				return
			}
			if ret.Code!=200{
				fmt.Printf("sendreward to php failed  rsp is :%v  \n", string(body))

			}
			return
		}
		fmt.Println(resp.Status)

	}
	ray:=RequestPhpAy{RequestPhp{Uid:"dfdfgdg",Sc:1223324}}
	b,_:=json.Marshal(ray)
	fmt.Println(string(b))
	ctx := context.Background()
	if conn,err:=ReqHttpConnP.Get(ctx);err!=nil{
		panic(err)
	}else {
		_, _, err := conn.Agent.(*gorequest.SuperAgent).Post("http://XXXX"). //取出以后需要强制转换成你放进去的类型
			Send(ray).
			EndBytes(fun)
		conn.Ret(ReqHttpConnP)
		if err!=nil|| ret.Code!=200{
			fmt.Println(err)
		}
	}



}
```



 **查看池子 使用情况 以及回收间隔**
 
func  Test_sendrewardConcurrent(t *testing.T)  {

	for i:=0; i<100;i++  {
		go Test_sendreward(t)
	}
	go func() {

		for{
			//ReqHttpConnP.m.Lock()
			fmt.Println("pool watcher ",ReqHttpConnP)   //查看池子 使用情况 以及回收间隔
			//ReqHttpConnP.m.Unlock()
			c:=time.NewTimer(time.Second*2)
			<-c.C
		}

	}()


	c:=time.NewTimer(time.Minute*2)
	<-c.C
}
```
