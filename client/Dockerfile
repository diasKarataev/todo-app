FROM node:18.18.0

WORKDIR /usr/src/app

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE 0209

CMD ["npm", "start"]