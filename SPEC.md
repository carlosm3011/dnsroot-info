# SPEC and RULES

## SPEC

We want to (periodically) check the status of the DNS rootservers.

We want to be inspired by the MTR multitraceroute tool.

The output is to be a text table, with all root servers ordered by letter and four columns:

- IPv4 address -> the ipv4 of the root server
- IPv4 Instance -> the result of the CH query that results in the actual instance name
- IPv6 address
- IPv6 Instance -> the result of the CH query that results in the actual instance name

### Example table

Fields should be justified for easy readability.

SRV | IPv4          | IPv4 Result | IPv6           | IPv6 Result
A   | 198.41.0.4    | "nnn1-lon8" | 2001:503:ba3e: | "nnn1-frmrs-3"
B   | 170.247.170.2 |"b4-fra".    | 2801:168:10::b |"b3-fra"

## Stack and rules 

Go (I want / need a self contained binary)
Use the libraries most appropiate for each thing (cli parsing, dns querying, etc)
I want a makefile for building the app.

