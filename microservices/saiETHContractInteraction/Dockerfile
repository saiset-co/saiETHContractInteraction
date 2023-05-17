FROM ubuntu

WORKDIR /srv

COPY ./build /srv/
RUN apt-get update && apt-get install wget -y

RUN chmod +x sai-eth-interaction
CMD ./sai-eth-interaction start


EXPOSE 8818
EXPOSE 8804
EXPOSE 8819
