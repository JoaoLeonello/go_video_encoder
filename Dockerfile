FROM golang:1.17-alpine

# Configurar repositórios para suporte ao Python 2.7
RUN echo "http://dl-cdn.alpinelinux.org/alpine/v3.12/main" > /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.12/community" >> /etc/apk/repositories

# Instalar dependências necessárias, incluindo Python 2.7
RUN apk add --update --no-cache python2 bash ffmpeg make unzip gcc g++ scons && \
    ln -sf /usr/bin/python2 /usr/bin/python

# Instalar Bento4
WORKDIR /tmp/bento4
ENV BENTO4_BASE_URL="http://zebulon.bok.net/Bento4/source/" \
    BENTO4_VERSION="1-5-0-615" \
    BENTO4_CHECKSUM="5378dbb374343bc274981d6e2ef93bce0851bda1" \
    BENTO4_PATH="/opt/bento4" \
    BENTO4_TYPE="SRC"

RUN wget -q ${BENTO4_BASE_URL}/Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip && \
    echo "${BENTO4_CHECKSUM}  Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip" | sha1sum -c && \
    mkdir -p ${BENTO4_PATH}/bin && \
    unzip Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip -d ${BENTO4_PATH} && \
    rm -rf Bento4-${BENTO4_TYPE}-${BENTO4_VERSION}.zip && \
    cd ${BENTO4_PATH} && \
    python2 $(which scons) -u build_config=Release target=x86_64-unknown-linux && \
    cp -R ${BENTO4_PATH}/Build/Targets/x86_64-unknown-linux/Release/* ${BENTO4_PATH}/bin


# Adicionar Bento4 ao PATH
ENV PATH="/opt/bento4/bin:$PATH"

# Depuração: Verificar os binários do Bento4
RUN ls -la /opt/bento4/bin && echo $PATH && mp4dash --version || echo "mp4dash not found"

# Configurar o ambiente de trabalho
WORKDIR /go/src
ENTRYPOINT ["tail", "-f", "/dev/null"]
