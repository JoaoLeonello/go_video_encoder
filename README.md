# Introdução

Este projeto tem como objetivo **converter vídeos** para um formato compatível com **streaming adaptativo** (por exemplo, MPEG-DASH). Dessa forma, os clientes conseguem começar a reproduzir rapidamente (sem precisar baixar o arquivo todo) e ajustar a qualidade do vídeo em tempo real conforme a largura de banda disponível.

### Por que converter arquivos para streaming?
1. **Início rápido**: O player pode reproduzir o vídeo imediatamente, pois ele carrega pequenos segmentos em vez de baixar o vídeo inteiro.  
2. **Streaming adaptativo**: O player escolhe automaticamente a qualidade ideal (resolução, bitrate) para o usuário, melhorando a experiência em redes instáveis.  
3. **Escalabilidade**: Com o vídeo fragmentado, é mais fácil distribuir os pedaços do arquivo em diferentes servidores ou CDNs.  
4. **Compatibilidade**: Vários players de vídeo (HTML5, VLC, players nativos) suportam protocolos como DASH ou HLS para transmitir o conteúdo com maior robustez.

---

# Go Routines e Lock/Mutex

No Go, a concorrência é tratada principalmente com **goroutines** e **channels**, fornecendo um modelo de concorrência seguro e simples:

- **Goroutines**  
  São funções leves que rodam concorrentemente.
  ```go
  go func() {
      // Código executado em paralelo
  }()
Channels
São mecanismos de comunicação seguro entre goroutines.

```go
Copy code
c := make(chan string)

// Envio
c <- "mensagem"

// Recebimento
msg := <-c
Lock/Mutex (sync.Mutex)
```

Se você precisa de acesso exclusivo a um recurso compartilhado (por exemplo, incrementando uma variável global), pode proteger a seção de código usando sync.Mutex:

```go
var mu sync.Mutex

mu.Lock()
sharedResource++
mu.Unlock()
```

No projeto, usamos goroutines para processar uploads simultâneos e channels para controlar o fluxo das tarefas. Caso algum recurso compartilhado precise de proteção, usamos um Mutex para evitar condições de corrida:

```go
func (vu *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
    in := make(chan int, concurrency)

    // Inicia 'concurrency' goroutines
    for i := 0; i < concurrency; i++ {
        go vu.uploadWorker(in, doneUpload)
    }

    // ...
    return nil
}
```
Esse padrão aumenta a performance e a escalabilidade, pois várias partes do processo podem acontecer em paralelo, respeitando limites de concorrência e evitando bloqueios desnecessários.
