FROM golang:1.17-alpine

# Configurar repositórios para suporte ao Python 2.7
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.12/main" > /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.12/community" >> /etc/apk/repositories

# Instalar dependências necessárias, incluindo Python 2.7
RUN apk add --update --no-cache python2 bash ffmpeg make unzip gcc g++ scons && \
    ln -sf /usr/bin/python2 /usr/bin/python

# Instalar Bento4 e garantir que mp4dash seja incluído
WORKDIR /tmp/bento4
ENV BENTO4_BASE_URL="http://zebulon.bok.net/Bento4/source/" \
    BENTO4_VERSION="1-6-0-641" \
    BENTO4_CHECKSUM="1a682b3d99b5d24429c6a9ad43ebddde4d80d84f" \
    BENTO4_PATH="/opt/bento4" \
    BENTO4_TYPE="SRC"

RUN wget -q ${BENTO4_BASE_URL}/Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip && \
    mkdir -p ${BENTO4_PATH}/bin && \
    unzip Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip -d ${BENTO4_PATH} && \
    rm -rf Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip && \
    cd ${BENTO4_PATH} && \
    python2 $(which scons) -u build_config=Release target=x86_64-unknown-linux && \
    cp -R ${BENTO4_PATH}/Build/Targets/x86_64-unknown-linux/Release/* ${BENTO4_PATH}/bin

# Adicionar Bento4 ao PATH
ENV PATH="/opt/bento4/bin:$PATH"

# Testar se o mp4dash está funcionando
RUN mp4dash --version || echo "Erro: mp4dash não encontrado"

# Instalar Delve
RUN go install github.com/go-delve/delve/cmd/dlv@v1.8.3
EXPOSE 40000

# Configurar o ambiente de trabalho
WORKDIR /go/src
ENTRYPOINT ["tail", "-f", "/dev/null"]
