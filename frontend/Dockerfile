# Stage 1: Build the React application
FROM node:20-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Pass build-time arguments to be baked into the static files
ARG REACT_APP_API_BASE_URL
ARG REACT_APP_API_TOKEN
ENV REACT_APP_API_BASE_URL=${REACT_APP_API_BASE_URL}
ENV REACT_APP_API_TOKEN=${REACT_APP_API_TOKEN}
RUN npm run build

# Stage 2: Serve the application using Nginx
FROM nginx:1.27-alpine

# Remove default Nginx welcome page
RUN rm /etc/nginx/conf.d/default.conf

# Copy our custom Nginx config
COPY nginx/nginx.conf /etc/nginx/conf.d

# Copy the built static files from the build stage
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
