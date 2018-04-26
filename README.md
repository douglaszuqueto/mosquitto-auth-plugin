# Mosquitto Auth Plugin - PostgreSQL

Plugin para Autenticação e Autorização de usuários no Mosquitto

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

### Integrando

## Referências

* [Mosquitto Auth Plug](https://github.com/jpmens/mosquitto-auth-plug)
* [Mosquitto Go Auth](https://github.com/iegomez/mosquitto-go-auth)
