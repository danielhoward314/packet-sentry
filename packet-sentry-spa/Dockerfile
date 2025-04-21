# build stage
FROM node:lts-alpine as build-stage

# Set the working directory
WORKDIR /app

# Copy package.json and package-lock.json first
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application source code
COPY . .

# Build the application
RUN npm run build

# production stage
FROM nginx:stable-alpine as production-stage

# Copy custom Nginx configuration if any
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the build output to the Nginx html directory
COPY --from=build-stage /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Command to run Nginx
CMD ["nginx", "-g", "daemon off;"]
