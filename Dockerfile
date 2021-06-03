FROM ubuntu

RUN apt-get update && apt-get install -y nodejs

WORKDIR /
ENV PATH="/bin/wabt:/bin/minifyjs/bin:/bin/comby:${PATH}"
ENV DATA_PATH="/data"

ADD ./dependencies/wabt /bin/wabt
ADD ./dependencies/minifyjs /bin/minifyjs
ADD ./dependencies/comby /bin/comby
ADD ./dist/main.out /app

ENTRYPOINT ["/app"]