FROM armhf/alpine:edge
# FROM alpine:edge

RUN apk add --no-cache postgresql
ENV PGDATA /data/db
RUN mkdir -p /data/db

RUN chown postgres ${PGDATA}

ENV LANG en_US.utf8

EXPOSE 5432

USER postgres
RUN initdb
RUN echo -e "listen_addresses = '*'\nwal_level = 'replica'\nmax_wal_senders = 10\n" >> ${PGDATA}/postgresql.conf \
	&& echo -e "\nhost all all 0.0.0.0/0 md5" >> ${PGDATA}/pg_hba.conf \
	&& echo -e "\nhost replication repl 0.0.0.0/0 md5" >> ${PGDATA}/pg_hba.conf

# create user repl password 'repl' replication login;
# create user app password 'app' superuser;
# create database app owner app encoding utf8;

VOLUME /data

CMD ["postgres"]

