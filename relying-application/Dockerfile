FROM node:23-alpine AS builder
WORKDIR /app

ARG VITE_ATTESTER_URL=http://localhost:8080
ARG VITE_VERIFIER_URL=http://localhost:8081
ENV VITE_ATTESTER_URL=$VITE_ATTESTER_URL
ENV VITE_VERIFIER_URL=$VITE_VERIFIER_URL

COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]