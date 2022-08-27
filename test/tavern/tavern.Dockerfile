# tavern.Dockerfile
FROM python:3.10-slim

ARG TAVERNVER
RUN pip3 install tavern==$TAVERNVER pytest-html

