// copyright Jusong Chen

package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
)

const maxUploadSize = 20 * 1024 * 1024 // 20 mb

//lepus is the server to implement all Lepus features
type lepus struct {
	router        *chi.Mux
	addr          string
	version       string
	receiveDir    string
	educatorNames []string
	httpSrv       *http.Server
}

var s lepus

//Start starts Lepus server
func Start(addr string, staticHomeDir string, srvVersion string, receiveDir string) {
	var educatorNames = []string{"蔡春耀", "蔡温榄", "蔡小华", "曹世才", "曾敏毅", "沈淑耀", "陈本培", "陈福音", "陈济兴", "陈金治", "陈群青", "陈细英", "陈燕华", "陈由溪", "陈玉珍", "陈长青", "陈志章", "陈宗辉", "傅德卿", "官尚武", "何天祥", "洪大锋", "胡永模", "胡永岳", "江英岩", "蒋永潮", "李粹玉", "练锡康", "林秉松", "林端川", "林芳", "林嘉坚", "林茂英", "林培坚", "林岫英", "林月心", "林占樟", "林昭英", "刘开天", "刘世煌", "刘永宾", "陆佩珰", "罗朝东", "毛玉珍", "毛祖辉", "毛祖瑜", "欧阳兴", "潘宝升", "潘家健", "潘世英", "潘孝平", "钱振恒", "石柏仁", "童美霞", "王福庆", "王光琳", "王丽华", "王美珠", "王其本", "王强", "王如", "王玉芳", "危金炎", "魏友义", "吴大樑", "吴齐练", "吴生基", "肖方尤", "肖光磊", "肖忠波", "许美钗", "严秀凤", "杨诚", "杨虹", "杨孝华", "杨义森", "叶菁", "余春华", "余冬生", "余世棣", "余天雨", "余伟然", "詹玉赐", "张殿", "张桂贞", "张家新", "张孔烺", "张荣治", "张世年", "张芝萱", "赵文婷", "郑春高", "郑国钦", "郑玉珠", "郑宗振", "周治河", "庄可明", "庄瑞发"}

	s = lepus{
		router:        chi.NewRouter(),
		addr:          addr,
		version:       srvVersion,
		receiveDir:    receiveDir,
		educatorNames: educatorNames,
	}
	s.routes(staticHomeDir)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s.httpSrv = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}

	// this channel is for graceful shutdown:
	// if we receive an error, we can send it here to notify the LepusServer to be stopped
	shutdown := make(chan struct{}, 1)
	go func() {
		err := s.httpSrv.ListenAndServe()
		if err != nil {
			shutdown <- struct{}{}
			log.Printf("%v", err)
		}
	}()
	log.Printf("The service is ready to listen and serve on %s.", s.httpSrv.Addr)

	select {
	case killSignal := <-interrupt:
		switch killSignal {
		case os.Interrupt:
			log.Print("Got SIGINT...")
		case syscall.SIGTERM:
			log.Print("Got SIGTERM...")
		}
	case <-shutdown:
		log.Printf("Get server shutdown request")
	}

	log.Print("The service is shutting down...")
	s.httpSrv.Shutdown(context.Background())
	log.Print("Done")
}

//Stop will stop the  Lepus app
func Stop() {
	s.httpSrv.Shutdown(context.Background())

}
