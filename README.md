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
