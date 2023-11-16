# The following Dockerfile is simply an EXAMPLE of what a builder dockerfile may look like.
# It should not be used in production anywhere, by anyone. It exists for testing purposes only.

# This Dockerfile is used as a base image for the container that provides the build routine for the specific agent.
FROM alpine:latest

# Save your royal.yaml into an upper level 'config' directory.
WORKDIR /config

# For this example, the royal.yaml is stored in and copied from the 'configs' folder, however it must be placed at the
# repository root for installation to succeed.
COPY royal.yaml ./royal.yaml