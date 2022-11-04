FROM node:18-alpine3.16 as compile
WORKDIR /usr/app
COPY package*.json ./
COPY tsconfig*.json ./
RUN npm clean-install
COPY . ./
RUN npm run build

FROM node:18-alpine3.16 as stripped
WORKDIR /usr/app
COPY --from=compile /usr/app/package*.json ./
COPY --from=compile /usr/app/dist ./
RUN npm install --only=production

FROM gcr.io/distroless/nodejs18-debian11
WORKDIR /usr/app
COPY --from=stripped /usr/app ./
USER 1000
CMD ["index.js"]
