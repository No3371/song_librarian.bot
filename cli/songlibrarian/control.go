package main

// import (
// 	"fmt"
// 	"net"

// 	"No3371.github.com/song_librarian.bot/logger"
// 	"github.com/valyala/gorpc"
// )


// func startControlInterface (port uint16) (interfaceClosed chan struct{}, sErr error){
// 	s := &gorpc.Server{
// 		// Accept clients on this TCP address.
// 		Addr: fmt.Sprintf(":%d", port),
	
// 		// Echo handler - just return back the message we received from the client
// 		Handler: func(clientAddr string, request interface{}) interface{} {
// 			log.Printf("Obtained request %+v from the client %s\n", request, clientAddr)
// 			return request
// 		},
// 	}

// 	if sErr = s.Serve(); sErr != nil {
// 		logger.Logger.Errorf("Failed to serve the rpc server: %v", sErr)
// 		return nil, sErr
// 	}
// }



