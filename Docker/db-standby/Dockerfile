FROM alpine:edge

RUN apk add --no-cache postgresql
ARG master_host="host=db-master port=5432"
ENV PGDATA /data/db
RUN mkdir -p /data/db

RUN chown postgres ${PGDATA}

ENV LANG en_US.utf8

EXPOSE 5432

USER postgres

RUN initdb

RUN echo -e "listen_addresses = '*'\nhot_standby = 'on'" >> ${PGDATA}/postgresql.conf \
	&& echo -e "standby_mode = 'on'\nprimary_conninfo = '${master_host} user=repl password=repl'" >> ${PGDATA}/recovery.conf \
	&& echo -e "\nhost all all 0.0.0.0/0 md5" >> ${PGDATA}/pg_hba.conf

VOLUME /data

CMD ["postgres"]

