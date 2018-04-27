# Mosquitto Auth Plugin - PostgreSQL

Plugin para Autenticação e Autorização de usuários no Mosquitto

## Índice

- [Introdução](#introducao)
- [Dependências](#dependencias)
- [Build](#build)
    - [Standalone](#standalone)
    - [Docker](#docker)
- [Configuração](#configuracao)
    - [Geral](#geral)
    - [PostgreSQL](#postgresql)
- [PostgreSQL](#postgresql)
    - [Instalar](#instalar)
    - [Configurar](#configurar)
    - [Estrutura](#estrutura)
        - [Script](#script)
- [Integrando](#integrando)


## Introdução

Este plugin é originado do projeto [Mosquitto Go Auth](https://github.com/iegomez/mosquitto-go-auth) com uma diferença do mesmo ser adaptado para atender demandas mais especificas em conjunto com o Broker MQTT Mosquitto.

## Dependências

As depedências serão instaladas de acordo com o cenário que você escolher. Portanto irei me basear na distro **Ubuntu**.

Em resumo você terá 2 dependências principais: 

* Go(golang)
* Ferramentas para compilação

## Build

### Standalone

* Arquitetura X64
* Arquitetura ARM (Raspberry PI)

### Docker

* Arquitetura X64

## Configuração

### Geral

### PostgreSQL


## PostgreSQL

Como o plugin é fortemente atrelado ao Banco de dados, se faz necessário a criação do ecossistema base para funcionamento - "podendo" ser adaptado de acordo com o cenário do projeto.

### Instalar

Para começar, você precisa ao mínimo ter o **PostgreSQL** instalado em sua máquina(desktop, docker, raspberry, servidor...). Recomendo também alguma interface gŕafica para manipulação do banco de dados - eu, particularmente estou usando o **PgAdmin4**.

### Configurar

Depois de ter instalado o serviço, não se esqueça de configurar algumas coisas que se faz necessário no postgres.

* Liberar acesso remoto;
* Definição de senha para o usuário default;

### Estrutura

![img](https://raw.githubusercontent.com/douglaszuqueto/mosquitto-auth-plugin/master/.github/mosquitto-auth-plugin.png)

### Script

Nos 2 próximos tópicos será mostrado a estrutura base referente ao que é necessário para o correto funcionamento no que tange a **Atenticação e Autorização** de usuários e seus devidos tópicos.

Fica a sua escolha qual **database** utilizar. Você pode colocar dentro do banco que já vem pré-criado - *postgres*, mas também pode criar um de sua preferência ou até mesmo embarcar no mesmo database de seu projeto.

Outra escolha opcional é a questão do **schema** a ser utilizado. Como padrão do postgres deixei no *public*, mas você pode criar um chamado **mqtt** por exemplo - em meu projeto, é esta estrutura que eu sigo. Deixo abaixo a estrutura que pretendo adotar no projeto [controle de acesso](https://github.com/douglaszuqueto/controle-de-acesso)

* Database: controle-de-acesso
    * Schema: mqtt
        * Table: user
        * Table: acl 

Caso mude o schema, não esqueça de estar mudando o schema nos scripts abaixo.

De *CREATE TABLE public."user"* por **CREATE TABLE seu_schema."user"**

#### Tabela user

```sql
-- Table: public."user"

-- DROP TABLE public."user";

CREATE TABLE public."user"
(
    id bigint NOT NULL DEFAULT nextval('test_user_id_seq'::regclass),
    username character varying(100) COLLATE pg_catalog."default" NOT NULL,
    password character varying(200) COLLATE pg_catalog."default" NOT NULL,
    is_admin boolean NOT NULL,
    CONSTRAINT test_user_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public."user"
    OWNER to postgres;
```

#### Tabela ACL

```sql
-- Table: public.acl

-- DROP TABLE public.acl;

CREATE TABLE public.acl
(
    id bigint NOT NULL DEFAULT nextval('test_acl_id_seq'::regclass),
    topic character varying(200) COLLATE pg_catalog."default" NOT NULL,
    rw integer NOT NULL,
    client_id character varying COLLATE pg_catalog."default",
    state smallint NOT NULL DEFAULT 1,
    id_user bigint,
    CONSTRAINT test_acl_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.acl
    OWNER to postgres;
```

## Integrando

Para ocorrer a devida integração, você terá de adicionar algumas linhas no arquivo de configuração do seu broker mosquitto e depois adequar os valores de acordo
com o cenário que você possui.

Segue abaixo uma tabela referente a todas opçoes possíveis de configuração:


| Option         		| default           |  Mandatory  | Meaning                  |
| -------------- 		| ----------------- | :---------: | ------------------------ |
| pg_host           | localhost         |             | hostname/address
| pg_port           | 5432              |             | TCP port
| pg_user           |                   |     Y       | username
| pg_password       |                   |     Y       | password
| pg_dbname         |                   |     Y       | database name
| pg_userquery      |                   |     Y       | SQL for users
| pg_superquery     |                   |     N       | SQL for superusers
| pg_aclquery       |                   |     N       | SQL for ACLs
| pg_sslmode        |     disable       |     N       | SSL/TLS mode.
| pg_sslcert        |                   |     N       | SSL/TLS Client Cert.
| pg_sslkey         |                   |     N       | SSL/TLS Client Cert. Key
| pg_sslrootcert    |                   |     N       | SSL/TLS Root Cert

**Observação:** Todas opções listadas acima levam o prefixo auth_opt_. Ou seja, cada opção deverá ficar no seginte formato: **auth_opt_pg_host**

## Referências

* [Mosquitto Auth Plug](https://github.com/jpmens/mosquitto-auth-plug)
* [Mosquitto Go Auth](https://github.com/iegomez/mosquitto-go-auth)
