FROM node:18-alpine

WORKDIR /opt/app
COPY ./ /opt/app
ENV REACT_APP_BACKEND_WS_URL=ws://backend.sosanh.lehoanglong.net
ENV PORT=80
RUN npm install -g serve && npm install && npm run build