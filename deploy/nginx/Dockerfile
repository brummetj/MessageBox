FROM ubuntu
RUN apt-get update && apt-get install -y nginx
RUN ln -sf /dev/stdout /var/log/nginx/access.log
RUN ln -sf /dev/stderr /var/log/nginx/error.log
RUN rm /etc/nginx/sites-enabled/default
EXPOSE 80 443

ARG CONF=mb-blue.conf
COPY ./$CONF /etc/nginx/conf.d/default.conf
CMD ["nginx", "-g", "daemon off;"]