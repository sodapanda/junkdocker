FROM archlinux
WORKDIR /
USER root
RUN pacman -Syy
RUN yes | pacman -S iproute2
RUN mkdir app
COPY entrypoint.sh /app
COPY gost /app
EXPOSE 8800
WORKDIR /app
ENTRYPOINT ["./entrypoint.sh"]