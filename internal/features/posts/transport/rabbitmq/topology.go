package posts_transport_rabbitmq

// fanout: exchange копирует каждое сообщение во ВСЕ привязанные очереди, routing key игнорируется.
// У каждого инстанса приложения своя очередь -> каждый инстанс получает каждое событие
// и раздаёт его своим WebSocket-клиентам.
//
//	POST /posts (инстанс A) --publish--> [xnet.posts.events (fanout)]
//	                                        |-> amq.gen-AAA -> консьюмер A -> хаб A -> браузеры на A
//	                                        |-> amq.gen-BBB -> консьюмер B -> хаб B -> браузеры на B
const PostsEventsExchange = "xnet.posts.events"
