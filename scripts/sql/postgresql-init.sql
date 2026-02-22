SELECT 'CREATE USER keycloak PASSWORD ''keycloak'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'keycloak')\gexec
SELECT 'CREATE DATABASE keycloak OWNER keycloak'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak')\gexec

SELECT 'CREATE USER local PASSWORD ''local'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'local')\gexec
SELECT 'CREATE DATABASE local OWNER local'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'local')\gexec

SELECT 'CREATE USER latest PASSWORD ''latest'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'latest')\gexec
SELECT 'CREATE DATABASE latest OWNER latest'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'latest')\gexec

SELECT 'CREATE USER preview PASSWORD ''preview'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'preview')\gexec
SELECT 'CREATE DATABASE preview OWNER preview'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'preview')\gexec

SELECT 'CREATE USER portal PASSWORD ''portal'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'portal')\gexec
SELECT 'CREATE DATABASE portal OWNER portal'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'portal')\gexec

SELECT 'CREATE USER testdb PASSWORD ''testdb'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'testdb')\gexec
SELECT 'CREATE DATABASE testdb OWNER testdb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'testdb')\gexec

SELECT 'CREATE USER migrate PASSWORD ''migrate'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'migrate')\gexec
SELECT 'CREATE DATABASE migrate OWNER migrate'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'migrate')\gexec

SELECT 'CREATE USER acceptance PASSWORD ''acceptance'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'acceptance')\gexec
-- SELECT 'CREATE DATABASE acceptance OWNER acceptance TEMPLATE "template0" LOCALE_PROVIDER "icu" ICU_LOCALE "und-u-ks-level2";'
SELECT 'CREATE DATABASE acceptance OWNER acceptance;'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'acceptance')\gexec

SELECT 'CREATE USER photoprism_01 PASSWORD ''photoprism_01'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'photoprism_01')\gexec
SELECT 'CREATE DATABASE photoprism_01 OWNER photoprism_01'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'photoprism_01')\gexec

SELECT 'CREATE USER photoprism_02 PASSWORD ''photoprism_02'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'photoprism_02')\gexec
SELECT 'CREATE DATABASE photoprism_02 OWNER photoprism_02'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'photoprism_02')\gexec

SELECT 'CREATE USER photoprism_03 PASSWORD ''photoprism_03'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'photoprism_03')\gexec
SELECT 'CREATE DATABASE photoprism_03 OWNER photoprism_03'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'photoprism_03')\gexec

SELECT 'CREATE USER photoprism_04 PASSWORD ''photoprism_04'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'photoprism_04')\gexec
SELECT 'CREATE DATABASE photoprism_04 OWNER photoprism_04'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'photoprism_04')\gexec

SELECT 'CREATE USER photoprism_05 PASSWORD ''photoprism_05'''
WHERE NOT EXISTS (SELECT FROM pg_user WHERE usename = 'photoprism_05')\gexec
SELECT 'CREATE DATABASE photoprism_05 OWNER photoprism_05'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'photoprism_05')\gexec

    \c keycloak keycloak;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 258 (class 1259 OID 17041)
-- Name: admin_event_entity; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.admin_event_entity (
                                           id character varying(36) NOT NULL,
                                           admin_event_time bigint,
                                           realm_id character varying(255),
                                           operation_type character varying(255),
                                           auth_realm_id character varying(255),
                                           auth_client_id character varying(255),
                                           auth_user_id character varying(255),
                                           ip_address character varying(255),
                                           resource_path character varying(2550),
                                           representation text,
                                           error character varying(255),
                                           resource_type character varying(64)
);


ALTER TABLE public.admin_event_entity OWNER TO keycloak;

--
-- TOC entry 287 (class 1259 OID 17484)
-- Name: associated_policy; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.associated_policy (
                                          policy_id character varying(36) NOT NULL,
                                          associated_policy_id character varying(36) NOT NULL
);


ALTER TABLE public.associated_policy OWNER TO keycloak;

--
-- TOC entry 261 (class 1259 OID 17056)
-- Name: authentication_execution; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.authentication_execution (
                                                 id character varying(36) NOT NULL,
                                                 alias character varying(255),
                                                 authenticator character varying(36),
                                                 realm_id character varying(36),
                                                 flow_id character varying(36),
                                                 requirement integer,
                                                 priority integer,
                                                 authenticator_flow boolean DEFAULT false NOT NULL,
                                                 auth_flow_id character varying(36),
                                                 auth_config character varying(36)
);


ALTER TABLE public.authentication_execution OWNER TO keycloak;

--
-- TOC entry 260 (class 1259 OID 17051)
-- Name: authentication_flow; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.authentication_flow (
                                            id character varying(36) NOT NULL,
                                            alias character varying(255),
                                            description character varying(255),
                                            realm_id character varying(36),
                                            provider_id character varying(36) DEFAULT 'basic-flow'::character varying NOT NULL,
                                            top_level boolean DEFAULT false NOT NULL,
                                            built_in boolean DEFAULT false NOT NULL
);


ALTER TABLE public.authentication_flow OWNER TO keycloak;

--
-- TOC entry 259 (class 1259 OID 17046)
-- Name: authenticator_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.authenticator_config (
                                             id character varying(36) NOT NULL,
                                             alias character varying(255),
                                             realm_id character varying(36)
);


ALTER TABLE public.authenticator_config OWNER TO keycloak;

--
-- TOC entry 262 (class 1259 OID 17061)
-- Name: authenticator_config_entry; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.authenticator_config_entry (
                                                   authenticator_id character varying(36) NOT NULL,
                                                   value text,
                                                   name character varying(255) NOT NULL
);


ALTER TABLE public.authenticator_config_entry OWNER TO keycloak;

--
-- TOC entry 288 (class 1259 OID 17499)
-- Name: broker_link; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.broker_link (
                                    identity_provider character varying(255) NOT NULL,
                                    storage_provider_id character varying(255),
                                    realm_id character varying(36) NOT NULL,
                                    broker_user_id character varying(255),
                                    broker_username character varying(255),
                                    token text,
                                    user_id character varying(255) NOT NULL
);


ALTER TABLE public.broker_link OWNER TO keycloak;

--
-- TOC entry 219 (class 1259 OID 16422)
-- Name: client; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client (
                               id character varying(36) NOT NULL,
                               enabled boolean DEFAULT false NOT NULL,
                               full_scope_allowed boolean DEFAULT false NOT NULL,
                               client_id character varying(255),
                               not_before integer,
                               public_client boolean DEFAULT false NOT NULL,
                               secret character varying(255),
                               base_url character varying(255),
                               bearer_only boolean DEFAULT false NOT NULL,
                               management_url character varying(255),
                               surrogate_auth_required boolean DEFAULT false NOT NULL,
                               realm_id character varying(36),
                               protocol character varying(255),
                               node_rereg_timeout integer DEFAULT 0,
                               frontchannel_logout boolean DEFAULT false NOT NULL,
                               consent_required boolean DEFAULT false NOT NULL,
                               name character varying(255),
                               service_accounts_enabled boolean DEFAULT false NOT NULL,
                               client_authenticator_type character varying(255),
                               root_url character varying(255),
                               description character varying(255),
                               registration_token character varying(255),
                               standard_flow_enabled boolean DEFAULT true NOT NULL,
                               implicit_flow_enabled boolean DEFAULT false NOT NULL,
                               direct_access_grants_enabled boolean DEFAULT false NOT NULL,
                               always_display_in_console boolean DEFAULT false NOT NULL
);


ALTER TABLE public.client OWNER TO keycloak;

--
-- TOC entry 242 (class 1259 OID 16780)
-- Name: client_attributes; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_attributes (
                                          client_id character varying(36) NOT NULL,
                                          name character varying(255) NOT NULL,
                                          value text
);


ALTER TABLE public.client_attributes OWNER TO keycloak;

--
-- TOC entry 299 (class 1259 OID 17748)
-- Name: client_auth_flow_bindings; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_auth_flow_bindings (
                                                  client_id character varying(36) NOT NULL,
                                                  flow_id character varying(36),
                                                  binding_name character varying(255) NOT NULL
);


ALTER TABLE public.client_auth_flow_bindings OWNER TO keycloak;

--
-- TOC entry 298 (class 1259 OID 17623)
-- Name: client_initial_access; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_initial_access (
                                              id character varying(36) NOT NULL,
                                              realm_id character varying(36) NOT NULL,
                                              "timestamp" integer,
                                              expiration integer,
                                              count integer,
                                              remaining_count integer
);


ALTER TABLE public.client_initial_access OWNER TO keycloak;

--
-- TOC entry 244 (class 1259 OID 16790)
-- Name: client_node_registrations; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_node_registrations (
                                                  client_id character varying(36) NOT NULL,
                                                  value integer,
                                                  name character varying(255) NOT NULL
);


ALTER TABLE public.client_node_registrations OWNER TO keycloak;

--
-- TOC entry 276 (class 1259 OID 17289)
-- Name: client_scope; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_scope (
                                     id character varying(36) NOT NULL,
                                     name character varying(255),
                                     realm_id character varying(36),
                                     description character varying(255),
                                     protocol character varying(255)
);


ALTER TABLE public.client_scope OWNER TO keycloak;

--
-- TOC entry 277 (class 1259 OID 17303)
-- Name: client_scope_attributes; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_scope_attributes (
                                                scope_id character varying(36) NOT NULL,
                                                value character varying(2048),
                                                name character varying(255) NOT NULL
);


ALTER TABLE public.client_scope_attributes OWNER TO keycloak;

--
-- TOC entry 300 (class 1259 OID 17789)
-- Name: client_scope_client; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_scope_client (
                                            client_id character varying(255) NOT NULL,
                                            scope_id character varying(255) NOT NULL,
                                            default_scope boolean DEFAULT false NOT NULL
);


ALTER TABLE public.client_scope_client OWNER TO keycloak;

--
-- TOC entry 278 (class 1259 OID 17308)
-- Name: client_scope_role_mapping; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_scope_role_mapping (
                                                  scope_id character varying(36) NOT NULL,
                                                  role_id character varying(36) NOT NULL
);


ALTER TABLE public.client_scope_role_mapping OWNER TO keycloak;

--
-- TOC entry 220 (class 1259 OID 16433)
-- Name: client_session; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_session (
                                       id character varying(36) NOT NULL,
                                       client_id character varying(36),
                                       redirect_uri character varying(255),
                                       state character varying(255),
                                       "timestamp" integer,
                                       session_id character varying(36),
                                       auth_method character varying(255),
                                       realm_id character varying(255),
                                       auth_user_id character varying(36),
                                       current_action character varying(36)
);


ALTER TABLE public.client_session OWNER TO keycloak;

--
-- TOC entry 265 (class 1259 OID 17079)
-- Name: client_session_auth_status; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_session_auth_status (
                                                   authenticator character varying(36) NOT NULL,
                                                   status integer,
                                                   client_session character varying(36) NOT NULL
);


ALTER TABLE public.client_session_auth_status OWNER TO keycloak;

--
-- TOC entry 243 (class 1259 OID 16785)
-- Name: client_session_note; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_session_note (
                                            name character varying(255) NOT NULL,
                                            value character varying(255),
                                            client_session character varying(36) NOT NULL
);


ALTER TABLE public.client_session_note OWNER TO keycloak;

--
-- TOC entry 257 (class 1259 OID 16963)
-- Name: client_session_prot_mapper; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_session_prot_mapper (
                                                   protocol_mapper_id character varying(36) NOT NULL,
                                                   client_session character varying(36) NOT NULL
);


ALTER TABLE public.client_session_prot_mapper OWNER TO keycloak;

--
-- TOC entry 221 (class 1259 OID 16438)
-- Name: client_session_role; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_session_role (
                                            role_id character varying(255) NOT NULL,
                                            client_session character varying(36) NOT NULL
);


ALTER TABLE public.client_session_role OWNER TO keycloak;

--
-- TOC entry 266 (class 1259 OID 17160)
-- Name: client_user_session_note; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.client_user_session_note (
                                                 name character varying(255) NOT NULL,
                                                 value character varying(2048),
                                                 client_session character varying(36) NOT NULL
);


ALTER TABLE public.client_user_session_note OWNER TO keycloak;

--
-- TOC entry 296 (class 1259 OID 17544)
-- Name: component; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.component (
                                  id character varying(36) NOT NULL,
                                  name character varying(255),
                                  parent_id character varying(36),
                                  provider_id character varying(36),
                                  provider_type character varying(255),
                                  realm_id character varying(36),
                                  sub_type character varying(255)
);


ALTER TABLE public.component OWNER TO keycloak;

--
-- TOC entry 295 (class 1259 OID 17539)
-- Name: component_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.component_config (
                                         id character varying(36) NOT NULL,
                                         component_id character varying(36) NOT NULL,
                                         name character varying(255) NOT NULL,
                                         value text
);


ALTER TABLE public.component_config OWNER TO keycloak;

--
-- TOC entry 222 (class 1259 OID 16441)
-- Name: composite_role; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.composite_role (
                                       composite character varying(36) NOT NULL,
                                       child_role character varying(36) NOT NULL
);


ALTER TABLE public.composite_role OWNER TO keycloak;

--
-- TOC entry 223 (class 1259 OID 16444)
-- Name: credential; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.credential (
                                   id character varying(36) NOT NULL,
                                   salt bytea,
                                   type character varying(255),
                                   user_id character varying(36),
                                   created_date bigint,
                                   user_label character varying(255),
                                   secret_data text,
                                   credential_data text,
                                   priority integer
);


ALTER TABLE public.credential OWNER TO keycloak;

--
-- TOC entry 218 (class 1259 OID 16414)
-- Name: databasechangelog; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.databasechangelog (
                                          id character varying(255) NOT NULL,
                                          author character varying(255) NOT NULL,
                                          filename character varying(255) NOT NULL,
                                          dateexecuted timestamp without time zone NOT NULL,
                                          orderexecuted integer NOT NULL,
                                          exectype character varying(10) NOT NULL,
                                          md5sum character varying(35),
                                          description character varying(255),
                                          comments character varying(255),
                                          tag character varying(255),
                                          liquibase character varying(20),
                                          contexts character varying(255),
                                          labels character varying(255),
                                          deployment_id character varying(10)
);


ALTER TABLE public.databasechangelog OWNER TO keycloak;

--
-- TOC entry 217 (class 1259 OID 16409)
-- Name: databasechangeloglock; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.databasechangeloglock (
                                              id integer NOT NULL,
                                              locked boolean NOT NULL,
                                              lockgranted timestamp without time zone,
                                              lockedby character varying(255)
);


ALTER TABLE public.databasechangeloglock OWNER TO keycloak;

--
-- TOC entry 301 (class 1259 OID 17805)
-- Name: default_client_scope; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.default_client_scope (
                                             realm_id character varying(36) NOT NULL,
                                             scope_id character varying(36) NOT NULL,
                                             default_scope boolean DEFAULT false NOT NULL
);


ALTER TABLE public.default_client_scope OWNER TO keycloak;

--
-- TOC entry 224 (class 1259 OID 16449)
-- Name: event_entity; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.event_entity (
                                     id character varying(36) NOT NULL,
                                     client_id character varying(255),
                                     details_json character varying(2550),
                                     error character varying(255),
                                     ip_address character varying(255),
                                     realm_id character varying(255),
                                     session_id character varying(255),
                                     event_time bigint,
                                     type character varying(255),
                                     user_id character varying(255),
                                     details_json_long_value text
);


ALTER TABLE public.event_entity OWNER TO keycloak;

--
-- TOC entry 289 (class 1259 OID 17504)
-- Name: fed_user_attribute; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_attribute (
                                           id character varying(36) NOT NULL,
                                           name character varying(255) NOT NULL,
                                           user_id character varying(255) NOT NULL,
                                           realm_id character varying(36) NOT NULL,
                                           storage_provider_id character varying(36),
                                           value character varying(2024),
                                           long_value_hash bytea,
                                           long_value_hash_lower_case bytea,
                                           long_value text
);


ALTER TABLE public.fed_user_attribute OWNER TO keycloak;

--
-- TOC entry 290 (class 1259 OID 17509)
-- Name: fed_user_consent; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_consent (
                                         id character varying(36) NOT NULL,
                                         client_id character varying(255),
                                         user_id character varying(255) NOT NULL,
                                         realm_id character varying(36) NOT NULL,
                                         storage_provider_id character varying(36),
                                         created_date bigint,
                                         last_updated_date bigint,
                                         client_storage_provider character varying(36),
                                         external_client_id character varying(255)
);


ALTER TABLE public.fed_user_consent OWNER TO keycloak;

--
-- TOC entry 303 (class 1259 OID 17831)
-- Name: fed_user_consent_cl_scope; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_consent_cl_scope (
                                                  user_consent_id character varying(36) NOT NULL,
                                                  scope_id character varying(36) NOT NULL
);


ALTER TABLE public.fed_user_consent_cl_scope OWNER TO keycloak;

--
-- TOC entry 291 (class 1259 OID 17518)
-- Name: fed_user_credential; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_credential (
                                            id character varying(36) NOT NULL,
                                            salt bytea,
                                            type character varying(255),
                                            created_date bigint,
                                            user_id character varying(255) NOT NULL,
                                            realm_id character varying(36) NOT NULL,
                                            storage_provider_id character varying(36),
                                            user_label character varying(255),
                                            secret_data text,
                                            credential_data text,
                                            priority integer
);


ALTER TABLE public.fed_user_credential OWNER TO keycloak;

--
-- TOC entry 292 (class 1259 OID 17527)
-- Name: fed_user_group_membership; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_group_membership (
                                                  group_id character varying(36) NOT NULL,
                                                  user_id character varying(255) NOT NULL,
                                                  realm_id character varying(36) NOT NULL,
                                                  storage_provider_id character varying(36)
);


ALTER TABLE public.fed_user_group_membership OWNER TO keycloak;

--
-- TOC entry 293 (class 1259 OID 17530)
-- Name: fed_user_required_action; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_required_action (
                                                 required_action character varying(255) DEFAULT ' '::character varying NOT NULL,
                                                 user_id character varying(255) NOT NULL,
                                                 realm_id character varying(36) NOT NULL,
                                                 storage_provider_id character varying(36)
);


ALTER TABLE public.fed_user_required_action OWNER TO keycloak;

--
-- TOC entry 294 (class 1259 OID 17536)
-- Name: fed_user_role_mapping; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.fed_user_role_mapping (
                                              role_id character varying(36) NOT NULL,
                                              user_id character varying(255) NOT NULL,
                                              realm_id character varying(36) NOT NULL,
                                              storage_provider_id character varying(36)
);


ALTER TABLE public.fed_user_role_mapping OWNER TO keycloak;

--
-- TOC entry 247 (class 1259 OID 16826)
-- Name: federated_identity; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.federated_identity (
                                           identity_provider character varying(255) NOT NULL,
                                           realm_id character varying(36),
                                           federated_user_id character varying(255),
                                           federated_username character varying(255),
                                           token text,
                                           user_id character varying(36) NOT NULL
);


ALTER TABLE public.federated_identity OWNER TO keycloak;

--
-- TOC entry 297 (class 1259 OID 17601)
-- Name: federated_user; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.federated_user (
                                       id character varying(255) NOT NULL,
                                       storage_provider_id character varying(255),
                                       realm_id character varying(36) NOT NULL
);


ALTER TABLE public.federated_user OWNER TO keycloak;

--
-- TOC entry 273 (class 1259 OID 17228)
-- Name: group_attribute; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.group_attribute (
                                        id character varying(36) DEFAULT 'sybase-needs-something-here'::character varying NOT NULL,
                                        name character varying(255) NOT NULL,
                                        value character varying(255),
                                        group_id character varying(36) NOT NULL
);


ALTER TABLE public.group_attribute OWNER TO keycloak;

--
-- TOC entry 272 (class 1259 OID 17225)
-- Name: group_role_mapping; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.group_role_mapping (
                                           role_id character varying(36) NOT NULL,
                                           group_id character varying(36) NOT NULL
);


ALTER TABLE public.group_role_mapping OWNER TO keycloak;

--
-- TOC entry 248 (class 1259 OID 16831)
-- Name: identity_provider; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.identity_provider (
                                          internal_id character varying(36) NOT NULL,
                                          enabled boolean DEFAULT false NOT NULL,
                                          provider_alias character varying(255),
                                          provider_id character varying(255),
                                          store_token boolean DEFAULT false NOT NULL,
                                          authenticate_by_default boolean DEFAULT false NOT NULL,
                                          realm_id character varying(36),
                                          add_token_role boolean DEFAULT true NOT NULL,
                                          trust_email boolean DEFAULT false NOT NULL,
                                          first_broker_login_flow_id character varying(36),
                                          post_broker_login_flow_id character varying(36),
                                          provider_display_name character varying(255),
                                          link_only boolean DEFAULT false NOT NULL
);


ALTER TABLE public.identity_provider OWNER TO keycloak;

--
-- TOC entry 249 (class 1259 OID 16840)
-- Name: identity_provider_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.identity_provider_config (
                                                 identity_provider_id character varying(36) NOT NULL,
                                                 value text,
                                                 name character varying(255) NOT NULL
);


ALTER TABLE public.identity_provider_config OWNER TO keycloak;

--
-- TOC entry 254 (class 1259 OID 16944)
-- Name: identity_provider_mapper; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.identity_provider_mapper (
                                                 id character varying(36) NOT NULL,
                                                 name character varying(255) NOT NULL,
                                                 idp_alias character varying(255) NOT NULL,
                                                 idp_mapper_name character varying(255) NOT NULL,
                                                 realm_id character varying(36) NOT NULL
);


ALTER TABLE public.identity_provider_mapper OWNER TO keycloak;

--
-- TOC entry 255 (class 1259 OID 16949)
-- Name: idp_mapper_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.idp_mapper_config (
                                          idp_mapper_id character varying(36) NOT NULL,
                                          value text,
                                          name character varying(255) NOT NULL
);


ALTER TABLE public.idp_mapper_config OWNER TO keycloak;

--
-- TOC entry 271 (class 1259 OID 17222)
-- Name: keycloak_group; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.keycloak_group (
                                       id character varying(36) NOT NULL,
                                       name character varying(255),
                                       parent_group character varying(36) NOT NULL,
                                       realm_id character varying(36)
);


ALTER TABLE public.keycloak_group OWNER TO keycloak;

--
-- TOC entry 225 (class 1259 OID 16457)
-- Name: keycloak_role; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.keycloak_role (
                                      id character varying(36) NOT NULL,
                                      client_realm_constraint character varying(255),
                                      client_role boolean DEFAULT false NOT NULL,
                                      description character varying(255),
                                      name character varying(255),
                                      realm_id character varying(255),
                                      client character varying(36),
                                      realm character varying(36)
);


ALTER TABLE public.keycloak_role OWNER TO keycloak;

--
-- TOC entry 253 (class 1259 OID 16941)
-- Name: migration_model; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.migration_model (
                                        id character varying(36) NOT NULL,
                                        version character varying(36),
                                        update_time bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.migration_model OWNER TO keycloak;

--
-- TOC entry 270 (class 1259 OID 17213)
-- Name: offline_client_session; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.offline_client_session (
                                               user_session_id character varying(36) NOT NULL,
                                               client_id character varying(255) NOT NULL,
                                               offline_flag character varying(4) NOT NULL,
                                               "timestamp" integer,
                                               data text,
                                               client_storage_provider character varying(36) DEFAULT 'local'::character varying NOT NULL,
                                               external_client_id character varying(255) DEFAULT 'local'::character varying NOT NULL,
                                               version integer DEFAULT 0
);


ALTER TABLE public.offline_client_session OWNER TO keycloak;

--
-- TOC entry 269 (class 1259 OID 17208)
-- Name: offline_user_session; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.offline_user_session (
                                             user_session_id character varying(36) NOT NULL,
                                             user_id character varying(255) NOT NULL,
                                             realm_id character varying(36) NOT NULL,
                                             created_on integer NOT NULL,
                                             offline_flag character varying(4) NOT NULL,
                                             data text,
                                             last_session_refresh integer DEFAULT 0 NOT NULL,
                                             broker_session_id character varying(1024),
                                             version integer DEFAULT 0
);


ALTER TABLE public.offline_user_session OWNER TO keycloak;

--
-- TOC entry 309 (class 1259 OID 17993)
-- Name: org; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.org (
                            id character varying(255) NOT NULL,
                            enabled boolean NOT NULL,
                            realm_id character varying(255) NOT NULL,
                            group_id character varying(255) NOT NULL,
                            name character varying(255) NOT NULL,
                            description character varying(4000)
);


ALTER TABLE public.org OWNER TO keycloak;

--
-- TOC entry 310 (class 1259 OID 18004)
-- Name: org_domain; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.org_domain (
                                   id character varying(36) NOT NULL,
                                   name character varying(255) NOT NULL,
                                   verified boolean NOT NULL,
                                   org_id character varying(255) NOT NULL
);


ALTER TABLE public.org_domain OWNER TO keycloak;

--
-- TOC entry 283 (class 1259 OID 17427)
-- Name: policy_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.policy_config (
                                      policy_id character varying(36) NOT NULL,
                                      name character varying(255) NOT NULL,
                                      value text
);


ALTER TABLE public.policy_config OWNER TO keycloak;

--
-- TOC entry 245 (class 1259 OID 16815)
-- Name: protocol_mapper; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.protocol_mapper (
                                        id character varying(36) NOT NULL,
                                        name character varying(255) NOT NULL,
                                        protocol character varying(255) NOT NULL,
                                        protocol_mapper_name character varying(255) NOT NULL,
                                        client_id character varying(36),
                                        client_scope_id character varying(36)
);


ALTER TABLE public.protocol_mapper OWNER TO keycloak;

--
-- TOC entry 246 (class 1259 OID 16821)
-- Name: protocol_mapper_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.protocol_mapper_config (
                                               protocol_mapper_id character varying(36) NOT NULL,
                                               value text,
                                               name character varying(255) NOT NULL
);


ALTER TABLE public.protocol_mapper_config OWNER TO keycloak;

--
-- TOC entry 226 (class 1259 OID 16463)
-- Name: realm; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm (
                              id character varying(36) NOT NULL,
                              access_code_lifespan integer,
                              user_action_lifespan integer,
                              access_token_lifespan integer,
                              account_theme character varying(255),
                              admin_theme character varying(255),
                              email_theme character varying(255),
                              enabled boolean DEFAULT false NOT NULL,
                              events_enabled boolean DEFAULT false NOT NULL,
                              events_expiration bigint,
                              login_theme character varying(255),
                              name character varying(255),
                              not_before integer,
                              password_policy character varying(2550),
                              registration_allowed boolean DEFAULT false NOT NULL,
                              remember_me boolean DEFAULT false NOT NULL,
                              reset_password_allowed boolean DEFAULT false NOT NULL,
                              social boolean DEFAULT false NOT NULL,
                              ssl_required character varying(255),
                              sso_idle_timeout integer,
                              sso_max_lifespan integer,
                              update_profile_on_soc_login boolean DEFAULT false NOT NULL,
                              verify_email boolean DEFAULT false NOT NULL,
                              master_admin_client character varying(36),
                              login_lifespan integer,
                              internationalization_enabled boolean DEFAULT false NOT NULL,
                              default_locale character varying(255),
                              reg_email_as_username boolean DEFAULT false NOT NULL,
                              admin_events_enabled boolean DEFAULT false NOT NULL,
                              admin_events_details_enabled boolean DEFAULT false NOT NULL,
                              edit_username_allowed boolean DEFAULT false NOT NULL,
                              otp_policy_counter integer DEFAULT 0,
                              otp_policy_window integer DEFAULT 1,
                              otp_policy_period integer DEFAULT 30,
                              otp_policy_digits integer DEFAULT 6,
                              otp_policy_alg character varying(36) DEFAULT 'HmacSHA1'::character varying,
                              otp_policy_type character varying(36) DEFAULT 'totp'::character varying,
                              browser_flow character varying(36),
                              registration_flow character varying(36),
                              direct_grant_flow character varying(36),
                              reset_credentials_flow character varying(36),
                              client_auth_flow character varying(36),
                              offline_session_idle_timeout integer DEFAULT 0,
                              revoke_refresh_token boolean DEFAULT false NOT NULL,
                              access_token_life_implicit integer DEFAULT 0,
                              login_with_email_allowed boolean DEFAULT true NOT NULL,
                              duplicate_emails_allowed boolean DEFAULT false NOT NULL,
                              docker_auth_flow character varying(36),
                              refresh_token_max_reuse integer DEFAULT 0,
                              allow_user_managed_access boolean DEFAULT false NOT NULL,
                              sso_max_lifespan_remember_me integer DEFAULT 0 NOT NULL,
                              sso_idle_timeout_remember_me integer DEFAULT 0 NOT NULL,
                              default_role character varying(255)
);


ALTER TABLE public.realm OWNER TO keycloak;

--
-- TOC entry 227 (class 1259 OID 16480)
-- Name: realm_attribute; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_attribute (
                                        name character varying(255) NOT NULL,
                                        realm_id character varying(36) NOT NULL,
                                        value text
);


ALTER TABLE public.realm_attribute OWNER TO keycloak;

--
-- TOC entry 275 (class 1259 OID 17237)
-- Name: realm_default_groups; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_default_groups (
                                             realm_id character varying(36) NOT NULL,
                                             group_id character varying(36) NOT NULL
);


ALTER TABLE public.realm_default_groups OWNER TO keycloak;

--
-- TOC entry 252 (class 1259 OID 16933)
-- Name: realm_enabled_event_types; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_enabled_event_types (
                                                  realm_id character varying(36) NOT NULL,
                                                  value character varying(255) NOT NULL
);


ALTER TABLE public.realm_enabled_event_types OWNER TO keycloak;

--
-- TOC entry 228 (class 1259 OID 16488)
-- Name: realm_events_listeners; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_events_listeners (
                                               realm_id character varying(36) NOT NULL,
                                               value character varying(255) NOT NULL
);


ALTER TABLE public.realm_events_listeners OWNER TO keycloak;

--
-- TOC entry 308 (class 1259 OID 17939)
-- Name: realm_localizations; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_localizations (
                                            realm_id character varying(255) NOT NULL,
                                            locale character varying(255) NOT NULL,
                                            texts text NOT NULL
);


ALTER TABLE public.realm_localizations OWNER TO keycloak;

--
-- TOC entry 229 (class 1259 OID 16491)
-- Name: realm_required_credential; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_required_credential (
                                                  type character varying(255) NOT NULL,
                                                  form_label character varying(255),
                                                  input boolean DEFAULT false NOT NULL,
                                                  secret boolean DEFAULT false NOT NULL,
                                                  realm_id character varying(36) NOT NULL
);


ALTER TABLE public.realm_required_credential OWNER TO keycloak;

--
-- TOC entry 230 (class 1259 OID 16498)
-- Name: realm_smtp_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_smtp_config (
                                          realm_id character varying(36) NOT NULL,
                                          value character varying(255),
                                          name character varying(255) NOT NULL
);


ALTER TABLE public.realm_smtp_config OWNER TO keycloak;

--
-- TOC entry 250 (class 1259 OID 16849)
-- Name: realm_supported_locales; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.realm_supported_locales (
                                                realm_id character varying(36) NOT NULL,
                                                value character varying(255) NOT NULL
);


ALTER TABLE public.realm_supported_locales OWNER TO keycloak;

--
-- TOC entry 231 (class 1259 OID 16508)
-- Name: redirect_uris; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.redirect_uris (
                                      client_id character varying(36) NOT NULL,
                                      value character varying(255) NOT NULL
);


ALTER TABLE public.redirect_uris OWNER TO keycloak;

--
-- TOC entry 268 (class 1259 OID 17172)
-- Name: required_action_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.required_action_config (
                                               required_action_id character varying(36) NOT NULL,
                                               value text,
                                               name character varying(255) NOT NULL
);


ALTER TABLE public.required_action_config OWNER TO keycloak;

--
-- TOC entry 267 (class 1259 OID 17165)
-- Name: required_action_provider; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.required_action_provider (
                                                 id character varying(36) NOT NULL,
                                                 alias character varying(255),
                                                 name character varying(255),
                                                 realm_id character varying(36),
                                                 enabled boolean DEFAULT false NOT NULL,
                                                 default_action boolean DEFAULT false NOT NULL,
                                                 provider_id character varying(255),
                                                 priority integer
);


ALTER TABLE public.required_action_provider OWNER TO keycloak;

--
-- TOC entry 305 (class 1259 OID 17870)
-- Name: resource_attribute; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_attribute (
                                           id character varying(36) DEFAULT 'sybase-needs-something-here'::character varying NOT NULL,
                                           name character varying(255) NOT NULL,
                                           value character varying(255),
                                           resource_id character varying(36) NOT NULL
);


ALTER TABLE public.resource_attribute OWNER TO keycloak;

--
-- TOC entry 285 (class 1259 OID 17454)
-- Name: resource_policy; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_policy (
                                        resource_id character varying(36) NOT NULL,
                                        policy_id character varying(36) NOT NULL
);


ALTER TABLE public.resource_policy OWNER TO keycloak;

--
-- TOC entry 284 (class 1259 OID 17439)
-- Name: resource_scope; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_scope (
                                       resource_id character varying(36) NOT NULL,
                                       scope_id character varying(36) NOT NULL
);


ALTER TABLE public.resource_scope OWNER TO keycloak;

--
-- TOC entry 279 (class 1259 OID 17377)
-- Name: resource_server; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_server (
                                        id character varying(36) NOT NULL,
                                        allow_rs_remote_mgmt boolean DEFAULT false NOT NULL,
                                        policy_enforce_mode smallint NOT NULL,
                                        decision_strategy smallint DEFAULT 1 NOT NULL
);


ALTER TABLE public.resource_server OWNER TO keycloak;

--
-- TOC entry 304 (class 1259 OID 17846)
-- Name: resource_server_perm_ticket; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_server_perm_ticket (
                                                    id character varying(36) NOT NULL,
                                                    owner character varying(255) NOT NULL,
                                                    requester character varying(255) NOT NULL,
                                                    created_timestamp bigint NOT NULL,
                                                    granted_timestamp bigint,
                                                    resource_id character varying(36) NOT NULL,
                                                    scope_id character varying(36),
                                                    resource_server_id character varying(36) NOT NULL,
                                                    policy_id character varying(36)
);


ALTER TABLE public.resource_server_perm_ticket OWNER TO keycloak;

--
-- TOC entry 282 (class 1259 OID 17413)
-- Name: resource_server_policy; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_server_policy (
                                               id character varying(36) NOT NULL,
                                               name character varying(255) NOT NULL,
                                               description character varying(255),
                                               type character varying(255) NOT NULL,
                                               decision_strategy smallint,
                                               logic smallint,
                                               resource_server_id character varying(36) NOT NULL,
                                               owner character varying(255)
);


ALTER TABLE public.resource_server_policy OWNER TO keycloak;

--
-- TOC entry 280 (class 1259 OID 17385)
-- Name: resource_server_resource; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_server_resource (
                                                 id character varying(36) NOT NULL,
                                                 name character varying(255) NOT NULL,
                                                 type character varying(255),
                                                 icon_uri character varying(255),
                                                 owner character varying(255) NOT NULL,
                                                 resource_server_id character varying(36) NOT NULL,
                                                 owner_managed_access boolean DEFAULT false NOT NULL,
                                                 display_name character varying(255)
);


ALTER TABLE public.resource_server_resource OWNER TO keycloak;

--
-- TOC entry 281 (class 1259 OID 17399)
-- Name: resource_server_scope; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_server_scope (
                                              id character varying(36) NOT NULL,
                                              name character varying(255) NOT NULL,
                                              icon_uri character varying(255),
                                              resource_server_id character varying(36) NOT NULL,
                                              display_name character varying(255)
);


ALTER TABLE public.resource_server_scope OWNER TO keycloak;

--
-- TOC entry 306 (class 1259 OID 17888)
-- Name: resource_uris; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.resource_uris (
                                      resource_id character varying(36) NOT NULL,
                                      value character varying(255) NOT NULL
);


ALTER TABLE public.resource_uris OWNER TO keycloak;

--
-- TOC entry 307 (class 1259 OID 17898)
-- Name: role_attribute; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.role_attribute (
                                       id character varying(36) NOT NULL,
                                       role_id character varying(36) NOT NULL,
                                       name character varying(255) NOT NULL,
                                       value character varying(255)
);


ALTER TABLE public.role_attribute OWNER TO keycloak;

--
-- TOC entry 232 (class 1259 OID 16511)
-- Name: scope_mapping; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.scope_mapping (
                                      client_id character varying(36) NOT NULL,
                                      role_id character varying(36) NOT NULL
);


ALTER TABLE public.scope_mapping OWNER TO keycloak;

--
-- TOC entry 286 (class 1259 OID 17469)
-- Name: scope_policy; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.scope_policy (
                                     scope_id character varying(36) NOT NULL,
                                     policy_id character varying(36) NOT NULL
);


ALTER TABLE public.scope_policy OWNER TO keycloak;

--
-- TOC entry 234 (class 1259 OID 16517)
-- Name: user_attribute; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_attribute (
                                       name character varying(255) NOT NULL,
                                       value character varying(255),
                                       user_id character varying(36) NOT NULL,
                                       id character varying(36) DEFAULT 'sybase-needs-something-here'::character varying NOT NULL,
                                       long_value_hash bytea,
                                       long_value_hash_lower_case bytea,
                                       long_value text
);


ALTER TABLE public.user_attribute OWNER TO keycloak;

--
-- TOC entry 256 (class 1259 OID 16954)
-- Name: user_consent; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_consent (
                                     id character varying(36) NOT NULL,
                                     client_id character varying(255),
                                     user_id character varying(36) NOT NULL,
                                     created_date bigint,
                                     last_updated_date bigint,
                                     client_storage_provider character varying(36),
                                     external_client_id character varying(255)
);


ALTER TABLE public.user_consent OWNER TO keycloak;

--
-- TOC entry 302 (class 1259 OID 17821)
-- Name: user_consent_client_scope; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_consent_client_scope (
                                                  user_consent_id character varying(36) NOT NULL,
                                                  scope_id character varying(36) NOT NULL
);


ALTER TABLE public.user_consent_client_scope OWNER TO keycloak;

--
-- TOC entry 235 (class 1259 OID 16522)
-- Name: user_entity; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_entity (
                                    id character varying(36) NOT NULL,
                                    email character varying(255),
                                    email_constraint character varying(255),
                                    email_verified boolean DEFAULT false NOT NULL,
                                    enabled boolean DEFAULT false NOT NULL,
                                    federation_link character varying(255),
                                    first_name character varying(255),
                                    last_name character varying(255),
                                    realm_id character varying(255),
                                    username character varying(255),
                                    created_timestamp bigint,
                                    service_account_client_link character varying(255),
                                    not_before integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.user_entity OWNER TO keycloak;

--
-- TOC entry 236 (class 1259 OID 16530)
-- Name: user_federation_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_federation_config (
                                               user_federation_provider_id character varying(36) NOT NULL,
                                               value character varying(255),
                                               name character varying(255) NOT NULL
);


ALTER TABLE public.user_federation_config OWNER TO keycloak;

--
-- TOC entry 263 (class 1259 OID 17066)
-- Name: user_federation_mapper; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_federation_mapper (
                                               id character varying(36) NOT NULL,
                                               name character varying(255) NOT NULL,
                                               federation_provider_id character varying(36) NOT NULL,
                                               federation_mapper_type character varying(255) NOT NULL,
                                               realm_id character varying(36) NOT NULL
);


ALTER TABLE public.user_federation_mapper OWNER TO keycloak;

--
-- TOC entry 264 (class 1259 OID 17071)
-- Name: user_federation_mapper_config; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_federation_mapper_config (
                                                      user_federation_mapper_id character varying(36) NOT NULL,
                                                      value character varying(255),
                                                      name character varying(255) NOT NULL
);


ALTER TABLE public.user_federation_mapper_config OWNER TO keycloak;

--
-- TOC entry 237 (class 1259 OID 16535)
-- Name: user_federation_provider; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_federation_provider (
                                                 id character varying(36) NOT NULL,
                                                 changed_sync_period integer,
                                                 display_name character varying(255),
                                                 full_sync_period integer,
                                                 last_sync integer,
                                                 priority integer,
                                                 provider_name character varying(255),
                                                 realm_id character varying(36)
);


ALTER TABLE public.user_federation_provider OWNER TO keycloak;

--
-- TOC entry 274 (class 1259 OID 17234)
-- Name: user_group_membership; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_group_membership (
                                              group_id character varying(36) NOT NULL,
                                              user_id character varying(36) NOT NULL
);


ALTER TABLE public.user_group_membership OWNER TO keycloak;

--
-- TOC entry 238 (class 1259 OID 16540)
-- Name: user_required_action; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_required_action (
                                             user_id character varying(36) NOT NULL,
                                             required_action character varying(255) DEFAULT ' '::character varying NOT NULL
);


ALTER TABLE public.user_required_action OWNER TO keycloak;

--
-- TOC entry 239 (class 1259 OID 16543)
-- Name: user_role_mapping; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_role_mapping (
                                          role_id character varying(255) NOT NULL,
                                          user_id character varying(36) NOT NULL
);


ALTER TABLE public.user_role_mapping OWNER TO keycloak;

--
-- TOC entry 240 (class 1259 OID 16546)
-- Name: user_session; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_session (
                                     id character varying(36) NOT NULL,
                                     auth_method character varying(255),
                                     ip_address character varying(255),
                                     last_session_refresh integer,
                                     login_username character varying(255),
                                     realm_id character varying(255),
                                     remember_me boolean DEFAULT false NOT NULL,
                                     started integer,
                                     user_id character varying(255),
                                     user_session_state integer,
                                     broker_session_id character varying(255),
                                     broker_user_id character varying(255)
);


ALTER TABLE public.user_session OWNER TO keycloak;

--
-- TOC entry 251 (class 1259 OID 16852)
-- Name: user_session_note; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.user_session_note (
                                          user_session character varying(36) NOT NULL,
                                          name character varying(255) NOT NULL,
                                          value character varying(2048)
);


ALTER TABLE public.user_session_note OWNER TO keycloak;

--
-- TOC entry 233 (class 1259 OID 16514)
-- Name: username_login_failure; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.username_login_failure (
                                               realm_id character varying(36) NOT NULL,
                                               username character varying(255) NOT NULL,
                                               failed_login_not_before integer,
                                               last_failure bigint,
                                               last_ip_failure character varying(255),
                                               num_failures integer
);


ALTER TABLE public.username_login_failure OWNER TO keycloak;

--
-- TOC entry 241 (class 1259 OID 16557)
-- Name: web_origins; Type: TABLE; Schema: public; Owner: keycloak
--

CREATE TABLE public.web_origins (
                                    client_id character varying(36) NOT NULL,
                                    value character varying(255) NOT NULL
);


ALTER TABLE public.web_origins OWNER TO keycloak;

--
-- TOC entry 3750 (class 2606 OID 17613)
-- Name: username_login_failure CONSTRAINT_17-2; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.username_login_failure
    ADD CONSTRAINT "CONSTRAINT_17-2" PRIMARY KEY (realm_id, username);


--
-- TOC entry 4008 (class 2606 OID 18010)
-- Name: org_domain ORG_DOMAIN_pkey; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.org_domain
    ADD CONSTRAINT "ORG_DOMAIN_pkey" PRIMARY KEY (id, name);


--
-- TOC entry 4002 (class 2606 OID 17999)
-- Name: org ORG_pkey; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.org
    ADD CONSTRAINT "ORG_pkey" PRIMARY KEY (id);


--
-- TOC entry 3723 (class 2606 OID 17922)
-- Name: keycloak_role UK_J3RWUVD56ONTGSUHOGM184WW2-2; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.keycloak_role
    ADD CONSTRAINT "UK_J3RWUVD56ONTGSUHOGM184WW2-2" UNIQUE (name, client_realm_constraint);


--
-- TOC entry 3971 (class 2606 OID 17752)
-- Name: client_auth_flow_bindings c_cli_flow_bind; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_auth_flow_bindings
    ADD CONSTRAINT c_cli_flow_bind PRIMARY KEY (client_id, binding_name);


--
-- TOC entry 3973 (class 2606 OID 17951)
-- Name: client_scope_client c_cli_scope_bind; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope_client
    ADD CONSTRAINT c_cli_scope_bind PRIMARY KEY (client_id, scope_id);


--
-- TOC entry 3968 (class 2606 OID 17627)
-- Name: client_initial_access cnstr_client_init_acc_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_initial_access
    ADD CONSTRAINT cnstr_client_init_acc_pk PRIMARY KEY (id);


--
-- TOC entry 3883 (class 2606 OID 17275)
-- Name: realm_default_groups con_group_id_def_groups; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_default_groups
    ADD CONSTRAINT con_group_id_def_groups UNIQUE (group_id);


--
-- TOC entry 3931 (class 2606 OID 17550)
-- Name: broker_link constr_broker_link_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.broker_link
    ADD CONSTRAINT constr_broker_link_pk PRIMARY KEY (identity_provider, user_id);


--
-- TOC entry 3854 (class 2606 OID 17184)
-- Name: client_user_session_note constr_cl_usr_ses_note; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_user_session_note
    ADD CONSTRAINT constr_cl_usr_ses_note PRIMARY KEY (client_session, name);


--
-- TOC entry 3959 (class 2606 OID 17570)
-- Name: component_config constr_component_config_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.component_config
    ADD CONSTRAINT constr_component_config_pk PRIMARY KEY (id);


--
-- TOC entry 3962 (class 2606 OID 17568)
-- Name: component constr_component_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.component
    ADD CONSTRAINT constr_component_pk PRIMARY KEY (id);


--
-- TOC entry 3951 (class 2606 OID 17566)
-- Name: fed_user_required_action constr_fed_required_action; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_required_action
    ADD CONSTRAINT constr_fed_required_action PRIMARY KEY (required_action, user_id);


--
-- TOC entry 3933 (class 2606 OID 17552)
-- Name: fed_user_attribute constr_fed_user_attr_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_attribute
    ADD CONSTRAINT constr_fed_user_attr_pk PRIMARY KEY (id);


--
-- TOC entry 3938 (class 2606 OID 17554)
-- Name: fed_user_consent constr_fed_user_consent_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_consent
    ADD CONSTRAINT constr_fed_user_consent_pk PRIMARY KEY (id);


--
-- TOC entry 3943 (class 2606 OID 17560)
-- Name: fed_user_credential constr_fed_user_cred_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_credential
    ADD CONSTRAINT constr_fed_user_cred_pk PRIMARY KEY (id);


--
-- TOC entry 3947 (class 2606 OID 17562)
-- Name: fed_user_group_membership constr_fed_user_group; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_group_membership
    ADD CONSTRAINT constr_fed_user_group PRIMARY KEY (group_id, user_id);


--
-- TOC entry 3955 (class 2606 OID 17564)
-- Name: fed_user_role_mapping constr_fed_user_role; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_role_mapping
    ADD CONSTRAINT constr_fed_user_role PRIMARY KEY (role_id, user_id);


--
-- TOC entry 3966 (class 2606 OID 17607)
-- Name: federated_user constr_federated_user; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.federated_user
    ADD CONSTRAINT constr_federated_user PRIMARY KEY (id);


--
-- TOC entry 3885 (class 2606 OID 17711)
-- Name: realm_default_groups constr_realm_default_groups; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_default_groups
    ADD CONSTRAINT constr_realm_default_groups PRIMARY KEY (realm_id, group_id);


--
-- TOC entry 3811 (class 2606 OID 17728)
-- Name: realm_enabled_event_types constr_realm_enabl_event_types; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_enabled_event_types
    ADD CONSTRAINT constr_realm_enabl_event_types PRIMARY KEY (realm_id, value);


--
-- TOC entry 3737 (class 2606 OID 17730)
-- Name: realm_events_listeners constr_realm_events_listeners; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_events_listeners
    ADD CONSTRAINT constr_realm_events_listeners PRIMARY KEY (realm_id, value);


--
-- TOC entry 3806 (class 2606 OID 17732)
-- Name: realm_supported_locales constr_realm_supported_locales; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_supported_locales
    ADD CONSTRAINT constr_realm_supported_locales PRIMARY KEY (realm_id, value);


--
-- TOC entry 3799 (class 2606 OID 16861)
-- Name: identity_provider constraint_2b; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider
    ADD CONSTRAINT constraint_2b PRIMARY KEY (internal_id);


--
-- TOC entry 3782 (class 2606 OID 16795)
-- Name: client_attributes constraint_3c; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_attributes
    ADD CONSTRAINT constraint_3c PRIMARY KEY (client_id, name);


--
-- TOC entry 3720 (class 2606 OID 16569)
-- Name: event_entity constraint_4; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.event_entity
    ADD CONSTRAINT constraint_4 PRIMARY KEY (id);


--
-- TOC entry 3795 (class 2606 OID 16863)
-- Name: federated_identity constraint_40; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.federated_identity
    ADD CONSTRAINT constraint_40 PRIMARY KEY (identity_provider, user_id);


--
-- TOC entry 3729 (class 2606 OID 16571)
-- Name: realm constraint_4a; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm
    ADD CONSTRAINT constraint_4a PRIMARY KEY (id);


--
-- TOC entry 3711 (class 2606 OID 16573)
-- Name: client_session_role constraint_5; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_role
    ADD CONSTRAINT constraint_5 PRIMARY KEY (client_session, role_id);


--
-- TOC entry 3777 (class 2606 OID 16575)
-- Name: user_session constraint_57; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_session
    ADD CONSTRAINT constraint_57 PRIMARY KEY (id);


--
-- TOC entry 3768 (class 2606 OID 16577)
-- Name: user_federation_provider constraint_5c; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_provider
    ADD CONSTRAINT constraint_5c PRIMARY KEY (id);


--
-- TOC entry 3785 (class 2606 OID 16797)
-- Name: client_session_note constraint_5e; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_note
    ADD CONSTRAINT constraint_5e PRIMARY KEY (client_session, name);


--
-- TOC entry 3703 (class 2606 OID 16581)
-- Name: client constraint_7; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client
    ADD CONSTRAINT constraint_7 PRIMARY KEY (id);


--
-- TOC entry 3708 (class 2606 OID 16583)
-- Name: client_session constraint_8; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session
    ADD CONSTRAINT constraint_8 PRIMARY KEY (id);


--
-- TOC entry 3747 (class 2606 OID 16585)
-- Name: scope_mapping constraint_81; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.scope_mapping
    ADD CONSTRAINT constraint_81 PRIMARY KEY (client_id, role_id);


--
-- TOC entry 3787 (class 2606 OID 16799)
-- Name: client_node_registrations constraint_84; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_node_registrations
    ADD CONSTRAINT constraint_84 PRIMARY KEY (client_id, name);


--
-- TOC entry 3734 (class 2606 OID 16587)
-- Name: realm_attribute constraint_9; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_attribute
    ADD CONSTRAINT constraint_9 PRIMARY KEY (name, realm_id);


--
-- TOC entry 3740 (class 2606 OID 16589)
-- Name: realm_required_credential constraint_92; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_required_credential
    ADD CONSTRAINT constraint_92 PRIMARY KEY (realm_id, type);


--
-- TOC entry 3725 (class 2606 OID 16591)
-- Name: keycloak_role constraint_a; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.keycloak_role
    ADD CONSTRAINT constraint_a PRIMARY KEY (id);


--
-- TOC entry 3831 (class 2606 OID 17715)
-- Name: admin_event_entity constraint_admin_event_entity; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.admin_event_entity
    ADD CONSTRAINT constraint_admin_event_entity PRIMARY KEY (id);


--
-- TOC entry 3844 (class 2606 OID 17092)
-- Name: authenticator_config_entry constraint_auth_cfg_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authenticator_config_entry
    ADD CONSTRAINT constraint_auth_cfg_pk PRIMARY KEY (authenticator_id, name);


--
-- TOC entry 3840 (class 2606 OID 17090)
-- Name: authentication_execution constraint_auth_exec_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authentication_execution
    ADD CONSTRAINT constraint_auth_exec_pk PRIMARY KEY (id);


--
-- TOC entry 3837 (class 2606 OID 17088)
-- Name: authentication_flow constraint_auth_flow_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authentication_flow
    ADD CONSTRAINT constraint_auth_flow_pk PRIMARY KEY (id);


--
-- TOC entry 3834 (class 2606 OID 17086)
-- Name: authenticator_config constraint_auth_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authenticator_config
    ADD CONSTRAINT constraint_auth_pk PRIMARY KEY (id);


--
-- TOC entry 3852 (class 2606 OID 17096)
-- Name: client_session_auth_status constraint_auth_status_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_auth_status
    ADD CONSTRAINT constraint_auth_status_pk PRIMARY KEY (client_session, authenticator);


--
-- TOC entry 3774 (class 2606 OID 16593)
-- Name: user_role_mapping constraint_c; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_role_mapping
    ADD CONSTRAINT constraint_c PRIMARY KEY (role_id, user_id);


--
-- TOC entry 3713 (class 2606 OID 17709)
-- Name: composite_role constraint_composite_role; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.composite_role
    ADD CONSTRAINT constraint_composite_role PRIMARY KEY (composite, child_role);


--
-- TOC entry 3829 (class 2606 OID 16979)
-- Name: client_session_prot_mapper constraint_cs_pmp_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_prot_mapper
    ADD CONSTRAINT constraint_cs_pmp_pk PRIMARY KEY (client_session, protocol_mapper_id);


--
-- TOC entry 3804 (class 2606 OID 16865)
-- Name: identity_provider_config constraint_d; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider_config
    ADD CONSTRAINT constraint_d PRIMARY KEY (identity_provider_id, name);


--
-- TOC entry 3917 (class 2606 OID 17433)
-- Name: policy_config constraint_dpc; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.policy_config
    ADD CONSTRAINT constraint_dpc PRIMARY KEY (policy_id, name);


--
-- TOC entry 3742 (class 2606 OID 16595)
-- Name: realm_smtp_config constraint_e; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_smtp_config
    ADD CONSTRAINT constraint_e PRIMARY KEY (realm_id, name);


--
-- TOC entry 3717 (class 2606 OID 16597)
-- Name: credential constraint_f; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.credential
    ADD CONSTRAINT constraint_f PRIMARY KEY (id);


--
-- TOC entry 3766 (class 2606 OID 16599)
-- Name: user_federation_config constraint_f9; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_config
    ADD CONSTRAINT constraint_f9 PRIMARY KEY (user_federation_provider_id, name);


--
-- TOC entry 3987 (class 2606 OID 17850)
-- Name: resource_server_perm_ticket constraint_fapmt; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_perm_ticket
    ADD CONSTRAINT constraint_fapmt PRIMARY KEY (id);


--
-- TOC entry 3902 (class 2606 OID 17391)
-- Name: resource_server_resource constraint_farsr; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_resource
    ADD CONSTRAINT constraint_farsr PRIMARY KEY (id);


--
-- TOC entry 3912 (class 2606 OID 17419)
-- Name: resource_server_policy constraint_farsrp; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_policy
    ADD CONSTRAINT constraint_farsrp PRIMARY KEY (id);


--
-- TOC entry 3928 (class 2606 OID 17488)
-- Name: associated_policy constraint_farsrpap; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.associated_policy
    ADD CONSTRAINT constraint_farsrpap PRIMARY KEY (policy_id, associated_policy_id);


--
-- TOC entry 3922 (class 2606 OID 17458)
-- Name: resource_policy constraint_farsrpp; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_policy
    ADD CONSTRAINT constraint_farsrpp PRIMARY KEY (resource_id, policy_id);


--
-- TOC entry 3907 (class 2606 OID 17405)
-- Name: resource_server_scope constraint_farsrs; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_scope
    ADD CONSTRAINT constraint_farsrs PRIMARY KEY (id);


--
-- TOC entry 3919 (class 2606 OID 17443)
-- Name: resource_scope constraint_farsrsp; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_scope
    ADD CONSTRAINT constraint_farsrsp PRIMARY KEY (resource_id, scope_id);


--
-- TOC entry 3925 (class 2606 OID 17473)
-- Name: scope_policy constraint_farsrsps; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.scope_policy
    ADD CONSTRAINT constraint_farsrsps PRIMARY KEY (scope_id, policy_id);


--
-- TOC entry 3758 (class 2606 OID 16601)
-- Name: user_entity constraint_fb; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_entity
    ADD CONSTRAINT constraint_fb PRIMARY KEY (id);


--
-- TOC entry 3850 (class 2606 OID 17100)
-- Name: user_federation_mapper_config constraint_fedmapper_cfg_pm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_mapper_config
    ADD CONSTRAINT constraint_fedmapper_cfg_pm PRIMARY KEY (user_federation_mapper_id, name);


--
-- TOC entry 3846 (class 2606 OID 17098)
-- Name: user_federation_mapper constraint_fedmapperpm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_mapper
    ADD CONSTRAINT constraint_fedmapperpm PRIMARY KEY (id);


--
-- TOC entry 3985 (class 2606 OID 17835)
-- Name: fed_user_consent_cl_scope constraint_fgrntcsnt_clsc_pm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.fed_user_consent_cl_scope
    ADD CONSTRAINT constraint_fgrntcsnt_clsc_pm PRIMARY KEY (user_consent_id, scope_id);


--
-- TOC entry 3981 (class 2606 OID 17825)
-- Name: user_consent_client_scope constraint_grntcsnt_clsc_pm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_consent_client_scope
    ADD CONSTRAINT constraint_grntcsnt_clsc_pm PRIMARY KEY (user_consent_id, scope_id);


--
-- TOC entry 3822 (class 2606 OID 16973)
-- Name: user_consent constraint_grntcsnt_pm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_consent
    ADD CONSTRAINT constraint_grntcsnt_pm PRIMARY KEY (id);


--
-- TOC entry 3869 (class 2606 OID 17242)
-- Name: keycloak_group constraint_group; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.keycloak_group
    ADD CONSTRAINT constraint_group PRIMARY KEY (id);


--
-- TOC entry 3876 (class 2606 OID 17249)
-- Name: group_attribute constraint_group_attribute_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.group_attribute
    ADD CONSTRAINT constraint_group_attribute_pk PRIMARY KEY (id);


--
-- TOC entry 3873 (class 2606 OID 17263)
-- Name: group_role_mapping constraint_group_role; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.group_role_mapping
    ADD CONSTRAINT constraint_group_role PRIMARY KEY (role_id, group_id);


--
-- TOC entry 3817 (class 2606 OID 16969)
-- Name: identity_provider_mapper constraint_idpm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider_mapper
    ADD CONSTRAINT constraint_idpm PRIMARY KEY (id);


--
-- TOC entry 3820 (class 2606 OID 17149)
-- Name: idp_mapper_config constraint_idpmconfig; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.idp_mapper_config
    ADD CONSTRAINT constraint_idpmconfig PRIMARY KEY (idp_mapper_id, name);


--
-- TOC entry 3814 (class 2606 OID 16967)
-- Name: migration_model constraint_migmod; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.migration_model
    ADD CONSTRAINT constraint_migmod PRIMARY KEY (id);


--
-- TOC entry 3866 (class 2606 OID 17928)
-- Name: offline_client_session constraint_offl_cl_ses_pk3; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.offline_client_session
    ADD CONSTRAINT constraint_offl_cl_ses_pk3 PRIMARY KEY (user_session_id, client_id, client_storage_provider, external_client_id, offline_flag);


--
-- TOC entry 3861 (class 2606 OID 17219)
-- Name: offline_user_session constraint_offl_us_ses_pk2; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.offline_user_session
    ADD CONSTRAINT constraint_offl_us_ses_pk2 PRIMARY KEY (user_session_id, offline_flag);


--
-- TOC entry 3789 (class 2606 OID 16859)
-- Name: protocol_mapper constraint_pcm; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.protocol_mapper
    ADD CONSTRAINT constraint_pcm PRIMARY KEY (id);


--
-- TOC entry 3793 (class 2606 OID 17142)
-- Name: protocol_mapper_config constraint_pmconfig; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.protocol_mapper_config
    ADD CONSTRAINT constraint_pmconfig PRIMARY KEY (protocol_mapper_id, name);


--
-- TOC entry 3744 (class 2606 OID 17734)
-- Name: redirect_uris constraint_redirect_uris; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.redirect_uris
    ADD CONSTRAINT constraint_redirect_uris PRIMARY KEY (client_id, value);


--
-- TOC entry 3859 (class 2606 OID 17182)
-- Name: required_action_config constraint_req_act_cfg_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.required_action_config
    ADD CONSTRAINT constraint_req_act_cfg_pk PRIMARY KEY (required_action_id, name);


--
-- TOC entry 3856 (class 2606 OID 17180)
-- Name: required_action_provider constraint_req_act_prv_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.required_action_provider
    ADD CONSTRAINT constraint_req_act_prv_pk PRIMARY KEY (id);


--
-- TOC entry 3771 (class 2606 OID 17094)
-- Name: user_required_action constraint_required_action; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_required_action
    ADD CONSTRAINT constraint_required_action PRIMARY KEY (required_action, user_id);


--
-- TOC entry 3995 (class 2606 OID 17897)
-- Name: resource_uris constraint_resour_uris_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_uris
    ADD CONSTRAINT constraint_resour_uris_pk PRIMARY KEY (resource_id, value);


--
-- TOC entry 3997 (class 2606 OID 17904)
-- Name: role_attribute constraint_role_attribute_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.role_attribute
    ADD CONSTRAINT constraint_role_attribute_pk PRIMARY KEY (id);


--
-- TOC entry 3752 (class 2606 OID 17178)
-- Name: user_attribute constraint_user_attribute_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_attribute
    ADD CONSTRAINT constraint_user_attribute_pk PRIMARY KEY (id);


--
-- TOC entry 3880 (class 2606 OID 17256)
-- Name: user_group_membership constraint_user_group; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_group_membership
    ADD CONSTRAINT constraint_user_group PRIMARY KEY (group_id, user_id);


--
-- TOC entry 3809 (class 2606 OID 16869)
-- Name: user_session_note constraint_usn_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_session_note
    ADD CONSTRAINT constraint_usn_pk PRIMARY KEY (user_session, name);


--
-- TOC entry 3779 (class 2606 OID 17736)
-- Name: web_origins constraint_web_origins; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.web_origins
    ADD CONSTRAINT constraint_web_origins PRIMARY KEY (client_id, value);


--
-- TOC entry 3701 (class 2606 OID 16413)
-- Name: databasechangeloglock databasechangeloglock_pkey; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.databasechangeloglock
    ADD CONSTRAINT databasechangeloglock_pkey PRIMARY KEY (id);


--
-- TOC entry 3894 (class 2606 OID 17359)
-- Name: client_scope_attributes pk_cl_tmpl_attr; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope_attributes
    ADD CONSTRAINT pk_cl_tmpl_attr PRIMARY KEY (scope_id, name);


--
-- TOC entry 3889 (class 2606 OID 17318)
-- Name: client_scope pk_cli_template; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope
    ADD CONSTRAINT pk_cli_template PRIMARY KEY (id);


--
-- TOC entry 3900 (class 2606 OID 17689)
-- Name: resource_server pk_resource_server; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server
    ADD CONSTRAINT pk_resource_server PRIMARY KEY (id);


--
-- TOC entry 3898 (class 2606 OID 17347)
-- Name: client_scope_role_mapping pk_template_scope; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope_role_mapping
    ADD CONSTRAINT pk_template_scope PRIMARY KEY (scope_id, role_id);


--
-- TOC entry 3979 (class 2606 OID 17810)
-- Name: default_client_scope r_def_cli_scope_bind; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.default_client_scope
    ADD CONSTRAINT r_def_cli_scope_bind PRIMARY KEY (realm_id, scope_id);


--
-- TOC entry 4000 (class 2606 OID 17945)
-- Name: realm_localizations realm_localizations_pkey; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_localizations
    ADD CONSTRAINT realm_localizations_pkey PRIMARY KEY (realm_id, locale);


--
-- TOC entry 3993 (class 2606 OID 17877)
-- Name: resource_attribute res_attr_pk; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_attribute
    ADD CONSTRAINT res_attr_pk PRIMARY KEY (id);


--
-- TOC entry 3871 (class 2606 OID 17619)
-- Name: keycloak_group sibling_names; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.keycloak_group
    ADD CONSTRAINT sibling_names UNIQUE (realm_id, parent_group, name);


--
-- TOC entry 3802 (class 2606 OID 16916)
-- Name: identity_provider uk_2daelwnibji49avxsrtuf6xj33; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider
    ADD CONSTRAINT uk_2daelwnibji49avxsrtuf6xj33 UNIQUE (provider_alias, realm_id);


--
-- TOC entry 3706 (class 2606 OID 16605)
-- Name: client uk_b71cjlbenv945rb6gcon438at; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client
    ADD CONSTRAINT uk_b71cjlbenv945rb6gcon438at UNIQUE (realm_id, client_id);


--
-- TOC entry 3891 (class 2606 OID 17763)
-- Name: client_scope uk_cli_scope; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope
    ADD CONSTRAINT uk_cli_scope UNIQUE (realm_id, name);


--
-- TOC entry 3762 (class 2606 OID 16609)
-- Name: user_entity uk_dykn684sl8up1crfei6eckhd7; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_entity
    ADD CONSTRAINT uk_dykn684sl8up1crfei6eckhd7 UNIQUE (realm_id, email_constraint);


--
-- TOC entry 3825 (class 2606 OID 18014)
-- Name: user_consent uk_external_consent; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_consent
    ADD CONSTRAINT uk_external_consent UNIQUE (client_storage_provider, external_client_id, user_id);


--
-- TOC entry 3905 (class 2606 OID 17936)
-- Name: resource_server_resource uk_frsr6t700s9v50bu18ws5ha6; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_resource
    ADD CONSTRAINT uk_frsr6t700s9v50bu18ws5ha6 UNIQUE (name, owner, resource_server_id);


--
-- TOC entry 3991 (class 2606 OID 17932)
-- Name: resource_server_perm_ticket uk_frsr6t700s9v50bu18ws5pmt; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_perm_ticket
    ADD CONSTRAINT uk_frsr6t700s9v50bu18ws5pmt UNIQUE (owner, requester, resource_server_id, resource_id, scope_id);


--
-- TOC entry 3915 (class 2606 OID 17680)
-- Name: resource_server_policy uk_frsrpt700s9v50bu18ws5ha6; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_policy
    ADD CONSTRAINT uk_frsrpt700s9v50bu18ws5ha6 UNIQUE (name, resource_server_id);


--
-- TOC entry 3910 (class 2606 OID 17684)
-- Name: resource_server_scope uk_frsrst700s9v50bu18ws5ha6; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_scope
    ADD CONSTRAINT uk_frsrst700s9v50bu18ws5ha6 UNIQUE (name, resource_server_id);


--
-- TOC entry 3827 (class 2606 OID 18012)
-- Name: user_consent uk_local_consent; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_consent
    ADD CONSTRAINT uk_local_consent UNIQUE (client_id, user_id);


--
-- TOC entry 4004 (class 2606 OID 18003)
-- Name: org uk_org_group; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.org
    ADD CONSTRAINT uk_org_group UNIQUE (group_id);


--
-- TOC entry 4006 (class 2606 OID 18001)
-- Name: org uk_org_name; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.org
    ADD CONSTRAINT uk_org_name UNIQUE (realm_id, name);


--
-- TOC entry 3732 (class 2606 OID 16617)
-- Name: realm uk_orvsdmla56612eaefiq6wl5oi; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm
    ADD CONSTRAINT uk_orvsdmla56612eaefiq6wl5oi UNIQUE (name);


--
-- TOC entry 3764 (class 2606 OID 17609)
-- Name: user_entity uk_ru8tt6t700s9v50bu18ws5ha6; Type: CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_entity
    ADD CONSTRAINT uk_ru8tt6t700s9v50bu18ws5ha6 UNIQUE (realm_id, username);


--
-- TOC entry 3934 (class 1259 OID 17985)
-- Name: fed_user_attr_long_values; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX fed_user_attr_long_values ON public.fed_user_attribute USING btree (long_value_hash, name);


--
-- TOC entry 3935 (class 1259 OID 17987)
-- Name: fed_user_attr_long_values_lower_case; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX fed_user_attr_long_values_lower_case ON public.fed_user_attribute USING btree (long_value_hash_lower_case, name);


--
-- TOC entry 3832 (class 1259 OID 17961)
-- Name: idx_admin_event_time; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_admin_event_time ON public.admin_event_entity USING btree (realm_id, admin_event_time);


--
-- TOC entry 3929 (class 1259 OID 17633)
-- Name: idx_assoc_pol_assoc_pol_id; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_assoc_pol_assoc_pol_id ON public.associated_policy USING btree (associated_policy_id);


--
-- TOC entry 3835 (class 1259 OID 17637)
-- Name: idx_auth_config_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_auth_config_realm ON public.authenticator_config USING btree (realm_id);


--
-- TOC entry 3841 (class 1259 OID 17635)
-- Name: idx_auth_exec_flow; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_auth_exec_flow ON public.authentication_execution USING btree (flow_id);


--
-- TOC entry 3842 (class 1259 OID 17634)
-- Name: idx_auth_exec_realm_flow; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_auth_exec_realm_flow ON public.authentication_execution USING btree (realm_id, flow_id);


--
-- TOC entry 3838 (class 1259 OID 17636)
-- Name: idx_auth_flow_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_auth_flow_realm ON public.authentication_flow USING btree (realm_id);


--
-- TOC entry 3974 (class 1259 OID 17952)
-- Name: idx_cl_clscope; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_cl_clscope ON public.client_scope_client USING btree (scope_id);


--
-- TOC entry 3783 (class 1259 OID 17988)
-- Name: idx_client_att_by_name_value; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_client_att_by_name_value ON public.client_attributes USING btree (name, substr(value, 1, 255));


--
-- TOC entry 3704 (class 1259 OID 17937)
-- Name: idx_client_id; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_client_id ON public.client USING btree (client_id);


--
-- TOC entry 3969 (class 1259 OID 17677)
-- Name: idx_client_init_acc_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_client_init_acc_realm ON public.client_initial_access USING btree (realm_id);


--
-- TOC entry 3709 (class 1259 OID 17641)
-- Name: idx_client_session_session; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_client_session_session ON public.client_session USING btree (session_id);


--
-- TOC entry 3892 (class 1259 OID 17840)
-- Name: idx_clscope_attrs; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_clscope_attrs ON public.client_scope_attributes USING btree (scope_id);


--
-- TOC entry 3975 (class 1259 OID 17949)
-- Name: idx_clscope_cl; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_clscope_cl ON public.client_scope_client USING btree (client_id);


--
-- TOC entry 3790 (class 1259 OID 17837)
-- Name: idx_clscope_protmap; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_clscope_protmap ON public.protocol_mapper USING btree (client_scope_id);


--
-- TOC entry 3895 (class 1259 OID 17838)
-- Name: idx_clscope_role; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_clscope_role ON public.client_scope_role_mapping USING btree (scope_id);


--
-- TOC entry 3960 (class 1259 OID 17643)
-- Name: idx_compo_config_compo; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_compo_config_compo ON public.component_config USING btree (component_id);


--
-- TOC entry 3963 (class 1259 OID 17911)
-- Name: idx_component_provider_type; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_component_provider_type ON public.component USING btree (provider_type);


--
-- TOC entry 3964 (class 1259 OID 17642)
-- Name: idx_component_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_component_realm ON public.component USING btree (realm_id);


--
-- TOC entry 3714 (class 1259 OID 17644)
-- Name: idx_composite; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_composite ON public.composite_role USING btree (composite);


--
-- TOC entry 3715 (class 1259 OID 17645)
-- Name: idx_composite_child; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_composite_child ON public.composite_role USING btree (child_role);


--
-- TOC entry 3976 (class 1259 OID 17843)
-- Name: idx_defcls_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_defcls_realm ON public.default_client_scope USING btree (realm_id);


--
-- TOC entry 3977 (class 1259 OID 17844)
-- Name: idx_defcls_scope; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_defcls_scope ON public.default_client_scope USING btree (scope_id);


--
-- TOC entry 3721 (class 1259 OID 17938)
-- Name: idx_event_time; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_event_time ON public.event_entity USING btree (realm_id, event_time);


--
-- TOC entry 3796 (class 1259 OID 17376)
-- Name: idx_fedidentity_feduser; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fedidentity_feduser ON public.federated_identity USING btree (federated_user_id);


--
-- TOC entry 3797 (class 1259 OID 17375)
-- Name: idx_fedidentity_user; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fedidentity_user ON public.federated_identity USING btree (user_id);


--
-- TOC entry 3936 (class 1259 OID 17737)
-- Name: idx_fu_attribute; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_attribute ON public.fed_user_attribute USING btree (user_id, realm_id, name);


--
-- TOC entry 3939 (class 1259 OID 17757)
-- Name: idx_fu_cnsnt_ext; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_cnsnt_ext ON public.fed_user_consent USING btree (user_id, client_storage_provider, external_client_id);


--
-- TOC entry 3940 (class 1259 OID 17920)
-- Name: idx_fu_consent; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_consent ON public.fed_user_consent USING btree (user_id, client_id);


--
-- TOC entry 3941 (class 1259 OID 17739)
-- Name: idx_fu_consent_ru; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_consent_ru ON public.fed_user_consent USING btree (realm_id, user_id);


--
-- TOC entry 3944 (class 1259 OID 17740)
-- Name: idx_fu_credential; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_credential ON public.fed_user_credential USING btree (user_id, type);


--
-- TOC entry 3945 (class 1259 OID 17741)
-- Name: idx_fu_credential_ru; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_credential_ru ON public.fed_user_credential USING btree (realm_id, user_id);


--
-- TOC entry 3948 (class 1259 OID 17742)
-- Name: idx_fu_group_membership; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_group_membership ON public.fed_user_group_membership USING btree (user_id, group_id);


--
-- TOC entry 3949 (class 1259 OID 17743)
-- Name: idx_fu_group_membership_ru; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_group_membership_ru ON public.fed_user_group_membership USING btree (realm_id, user_id);


--
-- TOC entry 3952 (class 1259 OID 17744)
-- Name: idx_fu_required_action; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_required_action ON public.fed_user_required_action USING btree (user_id, required_action);


--
-- TOC entry 3953 (class 1259 OID 17745)
-- Name: idx_fu_required_action_ru; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_required_action_ru ON public.fed_user_required_action USING btree (realm_id, user_id);


--
-- TOC entry 3956 (class 1259 OID 17746)
-- Name: idx_fu_role_mapping; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_role_mapping ON public.fed_user_role_mapping USING btree (user_id, role_id);


--
-- TOC entry 3957 (class 1259 OID 17747)
-- Name: idx_fu_role_mapping_ru; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_fu_role_mapping_ru ON public.fed_user_role_mapping USING btree (realm_id, user_id);


--
-- TOC entry 3877 (class 1259 OID 17963)
-- Name: idx_group_att_by_name_value; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_group_att_by_name_value ON public.group_attribute USING btree (name, ((value)::character varying(250)));


--
-- TOC entry 3878 (class 1259 OID 17648)
-- Name: idx_group_attr_group; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_group_attr_group ON public.group_attribute USING btree (group_id);


--
-- TOC entry 3874 (class 1259 OID 17649)
-- Name: idx_group_role_mapp_group; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_group_role_mapp_group ON public.group_role_mapping USING btree (group_id);


--
-- TOC entry 3818 (class 1259 OID 17651)
-- Name: idx_id_prov_mapp_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_id_prov_mapp_realm ON public.identity_provider_mapper USING btree (realm_id);


--
-- TOC entry 3800 (class 1259 OID 17650)
-- Name: idx_ident_prov_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_ident_prov_realm ON public.identity_provider USING btree (realm_id);


--
-- TOC entry 3726 (class 1259 OID 17652)
-- Name: idx_keycloak_role_client; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_keycloak_role_client ON public.keycloak_role USING btree (client);


--
-- TOC entry 3727 (class 1259 OID 17653)
-- Name: idx_keycloak_role_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_keycloak_role_realm ON public.keycloak_role USING btree (realm);


--
-- TOC entry 3862 (class 1259 OID 17992)
-- Name: idx_offline_uss_by_broker_session_id; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_offline_uss_by_broker_session_id ON public.offline_user_session USING btree (broker_session_id, realm_id);


--
-- TOC entry 3863 (class 1259 OID 17991)
-- Name: idx_offline_uss_by_last_session_refresh; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_offline_uss_by_last_session_refresh ON public.offline_user_session USING btree (realm_id, offline_flag, last_session_refresh);


--
-- TOC entry 3864 (class 1259 OID 17956)
-- Name: idx_offline_uss_by_user; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_offline_uss_by_user ON public.offline_user_session USING btree (user_id, realm_id, offline_flag);


--
-- TOC entry 3988 (class 1259 OID 18016)
-- Name: idx_perm_ticket_owner; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_perm_ticket_owner ON public.resource_server_perm_ticket USING btree (owner);


--
-- TOC entry 3989 (class 1259 OID 18015)
-- Name: idx_perm_ticket_requester; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_perm_ticket_requester ON public.resource_server_perm_ticket USING btree (requester);


--
-- TOC entry 3791 (class 1259 OID 17654)
-- Name: idx_protocol_mapper_client; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_protocol_mapper_client ON public.protocol_mapper USING btree (client_id);


--
-- TOC entry 3735 (class 1259 OID 17657)
-- Name: idx_realm_attr_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_attr_realm ON public.realm_attribute USING btree (realm_id);


--
-- TOC entry 3887 (class 1259 OID 17836)
-- Name: idx_realm_clscope; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_clscope ON public.client_scope USING btree (realm_id);


--
-- TOC entry 3886 (class 1259 OID 17658)
-- Name: idx_realm_def_grp_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_def_grp_realm ON public.realm_default_groups USING btree (realm_id);


--
-- TOC entry 3738 (class 1259 OID 17661)
-- Name: idx_realm_evt_list_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_evt_list_realm ON public.realm_events_listeners USING btree (realm_id);


--
-- TOC entry 3812 (class 1259 OID 17660)
-- Name: idx_realm_evt_types_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_evt_types_realm ON public.realm_enabled_event_types USING btree (realm_id);


--
-- TOC entry 3730 (class 1259 OID 17656)
-- Name: idx_realm_master_adm_cli; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_master_adm_cli ON public.realm USING btree (master_admin_client);


--
-- TOC entry 3807 (class 1259 OID 17662)
-- Name: idx_realm_supp_local_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_realm_supp_local_realm ON public.realm_supported_locales USING btree (realm_id);


--
-- TOC entry 3745 (class 1259 OID 17663)
-- Name: idx_redir_uri_client; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_redir_uri_client ON public.redirect_uris USING btree (client_id);


--
-- TOC entry 3857 (class 1259 OID 17664)
-- Name: idx_req_act_prov_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_req_act_prov_realm ON public.required_action_provider USING btree (realm_id);


--
-- TOC entry 3923 (class 1259 OID 17665)
-- Name: idx_res_policy_policy; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_res_policy_policy ON public.resource_policy USING btree (policy_id);


--
-- TOC entry 3920 (class 1259 OID 17666)
-- Name: idx_res_scope_scope; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_res_scope_scope ON public.resource_scope USING btree (scope_id);


--
-- TOC entry 3913 (class 1259 OID 17685)
-- Name: idx_res_serv_pol_res_serv; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_res_serv_pol_res_serv ON public.resource_server_policy USING btree (resource_server_id);


--
-- TOC entry 3903 (class 1259 OID 17686)
-- Name: idx_res_srv_res_res_srv; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_res_srv_res_res_srv ON public.resource_server_resource USING btree (resource_server_id);


--
-- TOC entry 3908 (class 1259 OID 17687)
-- Name: idx_res_srv_scope_res_srv; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_res_srv_scope_res_srv ON public.resource_server_scope USING btree (resource_server_id);


--
-- TOC entry 3998 (class 1259 OID 17910)
-- Name: idx_role_attribute; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_role_attribute ON public.role_attribute USING btree (role_id);


--
-- TOC entry 3896 (class 1259 OID 17839)
-- Name: idx_role_clscope; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_role_clscope ON public.client_scope_role_mapping USING btree (role_id);


--
-- TOC entry 3748 (class 1259 OID 17670)
-- Name: idx_scope_mapping_role; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_scope_mapping_role ON public.scope_mapping USING btree (role_id);


--
-- TOC entry 3926 (class 1259 OID 17671)
-- Name: idx_scope_policy_policy; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_scope_policy_policy ON public.scope_policy USING btree (policy_id);


--
-- TOC entry 3815 (class 1259 OID 17918)
-- Name: idx_update_time; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_update_time ON public.migration_model USING btree (update_time);


--
-- TOC entry 3867 (class 1259 OID 17365)
-- Name: idx_us_sess_id_on_cl_sess; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_us_sess_id_on_cl_sess ON public.offline_client_session USING btree (user_session_id);


--
-- TOC entry 3982 (class 1259 OID 17845)
-- Name: idx_usconsent_clscope; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_usconsent_clscope ON public.user_consent_client_scope USING btree (user_consent_id);


--
-- TOC entry 3983 (class 1259 OID 17962)
-- Name: idx_usconsent_scope_id; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_usconsent_scope_id ON public.user_consent_client_scope USING btree (scope_id);


--
-- TOC entry 3753 (class 1259 OID 17372)
-- Name: idx_user_attribute; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_attribute ON public.user_attribute USING btree (user_id);


--
-- TOC entry 3754 (class 1259 OID 17959)
-- Name: idx_user_attribute_name; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_attribute_name ON public.user_attribute USING btree (name, value);


--
-- TOC entry 3823 (class 1259 OID 17369)
-- Name: idx_user_consent; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_consent ON public.user_consent USING btree (user_id);


--
-- TOC entry 3718 (class 1259 OID 17373)
-- Name: idx_user_credential; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_credential ON public.credential USING btree (user_id);


--
-- TOC entry 3759 (class 1259 OID 17366)
-- Name: idx_user_email; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_email ON public.user_entity USING btree (email);


--
-- TOC entry 3881 (class 1259 OID 17368)
-- Name: idx_user_group_mapping; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_group_mapping ON public.user_group_membership USING btree (user_id);


--
-- TOC entry 3772 (class 1259 OID 17374)
-- Name: idx_user_reqactions; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_reqactions ON public.user_required_action USING btree (user_id);


--
-- TOC entry 3775 (class 1259 OID 17367)
-- Name: idx_user_role_mapping; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_role_mapping ON public.user_role_mapping USING btree (user_id);


--
-- TOC entry 3760 (class 1259 OID 17960)
-- Name: idx_user_service_account; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_user_service_account ON public.user_entity USING btree (realm_id, service_account_client_link);


--
-- TOC entry 3847 (class 1259 OID 17673)
-- Name: idx_usr_fed_map_fed_prv; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_usr_fed_map_fed_prv ON public.user_federation_mapper USING btree (federation_provider_id);


--
-- TOC entry 3848 (class 1259 OID 17674)
-- Name: idx_usr_fed_map_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_usr_fed_map_realm ON public.user_federation_mapper USING btree (realm_id);


--
-- TOC entry 3769 (class 1259 OID 17675)
-- Name: idx_usr_fed_prv_realm; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_usr_fed_prv_realm ON public.user_federation_provider USING btree (realm_id);


--
-- TOC entry 3780 (class 1259 OID 17676)
-- Name: idx_web_orig_client; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX idx_web_orig_client ON public.web_origins USING btree (client_id);


--
-- TOC entry 3755 (class 1259 OID 17984)
-- Name: user_attr_long_values; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX user_attr_long_values ON public.user_attribute USING btree (long_value_hash, name);


--
-- TOC entry 3756 (class 1259 OID 17986)
-- Name: user_attr_long_values_lower_case; Type: INDEX; Schema: public; Owner: keycloak
--

CREATE INDEX user_attr_long_values_lower_case ON public.user_attribute USING btree (long_value_hash_lower_case, name);


--
-- TOC entry 4050 (class 2606 OID 17101)
-- Name: client_session_auth_status auth_status_constraint; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_auth_status
    ADD CONSTRAINT auth_status_constraint FOREIGN KEY (client_session) REFERENCES public.client_session(id);


--
-- TOC entry 4034 (class 2606 OID 16870)
-- Name: identity_provider fk2b4ebc52ae5c3b34; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider
    ADD CONSTRAINT fk2b4ebc52ae5c3b34 FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4027 (class 2606 OID 16800)
-- Name: client_attributes fk3c47c64beacca966; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_attributes
    ADD CONSTRAINT fk3c47c64beacca966 FOREIGN KEY (client_id) REFERENCES public.client(id);


--
-- TOC entry 4033 (class 2606 OID 16880)
-- Name: federated_identity fk404288b92ef007a6; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.federated_identity
    ADD CONSTRAINT fk404288b92ef007a6 FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4029 (class 2606 OID 17027)
-- Name: client_node_registrations fk4129723ba992f594; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_node_registrations
    ADD CONSTRAINT fk4129723ba992f594 FOREIGN KEY (client_id) REFERENCES public.client(id);


--
-- TOC entry 4028 (class 2606 OID 16805)
-- Name: client_session_note fk5edfb00ff51c2736; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_note
    ADD CONSTRAINT fk5edfb00ff51c2736 FOREIGN KEY (client_session) REFERENCES public.client_session(id);


--
-- TOC entry 4037 (class 2606 OID 16910)
-- Name: user_session_note fk5edfb00ff51d3472; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_session_note
    ADD CONSTRAINT fk5edfb00ff51d3472 FOREIGN KEY (user_session) REFERENCES public.user_session(id);


--
-- TOC entry 4010 (class 2606 OID 16620)
-- Name: client_session_role fk_11b7sgqw18i532811v7o2dv76; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_role
    ADD CONSTRAINT fk_11b7sgqw18i532811v7o2dv76 FOREIGN KEY (client_session) REFERENCES public.client_session(id);


--
-- TOC entry 4019 (class 2606 OID 16625)
-- Name: redirect_uris fk_1burs8pb4ouj97h5wuppahv9f; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.redirect_uris
    ADD CONSTRAINT fk_1burs8pb4ouj97h5wuppahv9f FOREIGN KEY (client_id) REFERENCES public.client(id);


--
-- TOC entry 4023 (class 2606 OID 16630)
-- Name: user_federation_provider fk_1fj32f6ptolw2qy60cd8n01e8; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_provider
    ADD CONSTRAINT fk_1fj32f6ptolw2qy60cd8n01e8 FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4042 (class 2606 OID 17005)
-- Name: client_session_prot_mapper fk_33a8sgqw18i532811v7o2dk89; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session_prot_mapper
    ADD CONSTRAINT fk_33a8sgqw18i532811v7o2dk89 FOREIGN KEY (client_session) REFERENCES public.client_session(id);


--
-- TOC entry 4017 (class 2606 OID 16640)
-- Name: realm_required_credential fk_5hg65lybevavkqfki3kponh9v; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_required_credential
    ADD CONSTRAINT fk_5hg65lybevavkqfki3kponh9v FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4080 (class 2606 OID 17878)
-- Name: resource_attribute fk_5hrm2vlf9ql5fu022kqepovbr; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_attribute
    ADD CONSTRAINT fk_5hrm2vlf9ql5fu022kqepovbr FOREIGN KEY (resource_id) REFERENCES public.resource_server_resource(id);


--
-- TOC entry 4021 (class 2606 OID 16645)
-- Name: user_attribute fk_5hrm2vlf9ql5fu043kqepovbr; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_attribute
    ADD CONSTRAINT fk_5hrm2vlf9ql5fu043kqepovbr FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4024 (class 2606 OID 16655)
-- Name: user_required_action fk_6qj3w1jw9cvafhe19bwsiuvmd; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_required_action
    ADD CONSTRAINT fk_6qj3w1jw9cvafhe19bwsiuvmd FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4014 (class 2606 OID 16660)
-- Name: keycloak_role fk_6vyqfe4cn4wlq8r6kt5vdsj5c; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.keycloak_role
    ADD CONSTRAINT fk_6vyqfe4cn4wlq8r6kt5vdsj5c FOREIGN KEY (realm) REFERENCES public.realm(id);


--
-- TOC entry 4018 (class 2606 OID 16665)
-- Name: realm_smtp_config fk_70ej8xdxgxd0b9hh6180irr0o; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_smtp_config
    ADD CONSTRAINT fk_70ej8xdxgxd0b9hh6180irr0o FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4015 (class 2606 OID 16680)
-- Name: realm_attribute fk_8shxd6l3e9atqukacxgpffptw; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_attribute
    ADD CONSTRAINT fk_8shxd6l3e9atqukacxgpffptw FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4011 (class 2606 OID 16685)
-- Name: composite_role fk_a63wvekftu8jo1pnj81e7mce2; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.composite_role
    ADD CONSTRAINT fk_a63wvekftu8jo1pnj81e7mce2 FOREIGN KEY (composite) REFERENCES public.keycloak_role(id);


--
-- TOC entry 4045 (class 2606 OID 17121)
-- Name: authentication_execution fk_auth_exec_flow; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authentication_execution
    ADD CONSTRAINT fk_auth_exec_flow FOREIGN KEY (flow_id) REFERENCES public.authentication_flow(id);


--
-- TOC entry 4046 (class 2606 OID 17116)
-- Name: authentication_execution fk_auth_exec_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authentication_execution
    ADD CONSTRAINT fk_auth_exec_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4044 (class 2606 OID 17111)
-- Name: authentication_flow fk_auth_flow_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authentication_flow
    ADD CONSTRAINT fk_auth_flow_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4043 (class 2606 OID 17106)
-- Name: authenticator_config fk_auth_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.authenticator_config
    ADD CONSTRAINT fk_auth_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4009 (class 2606 OID 16690)
-- Name: client_session fk_b4ao2vcvat6ukau74wbwtfqo1; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_session
    ADD CONSTRAINT fk_b4ao2vcvat6ukau74wbwtfqo1 FOREIGN KEY (session_id) REFERENCES public.user_session(id);


--
-- TOC entry 4025 (class 2606 OID 16695)
-- Name: user_role_mapping fk_c4fqv34p1mbylloxang7b1q3l; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_role_mapping
    ADD CONSTRAINT fk_c4fqv34p1mbylloxang7b1q3l FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4057 (class 2606 OID 17784)
-- Name: client_scope_attributes fk_cl_scope_attr_scope; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope_attributes
    ADD CONSTRAINT fk_cl_scope_attr_scope FOREIGN KEY (scope_id) REFERENCES public.client_scope(id);


--
-- TOC entry 4058 (class 2606 OID 17774)
-- Name: client_scope_role_mapping fk_cl_scope_rm_scope; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_scope_role_mapping
    ADD CONSTRAINT fk_cl_scope_rm_scope FOREIGN KEY (scope_id) REFERENCES public.client_scope(id);


--
-- TOC entry 4051 (class 2606 OID 17190)
-- Name: client_user_session_note fk_cl_usr_ses_note; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_user_session_note
    ADD CONSTRAINT fk_cl_usr_ses_note FOREIGN KEY (client_session) REFERENCES public.client_session(id);


--
-- TOC entry 4030 (class 2606 OID 17769)
-- Name: protocol_mapper fk_cli_scope_mapper; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.protocol_mapper
    ADD CONSTRAINT fk_cli_scope_mapper FOREIGN KEY (client_scope_id) REFERENCES public.client_scope(id);


--
-- TOC entry 4073 (class 2606 OID 17628)
-- Name: client_initial_access fk_client_init_acc_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.client_initial_access
    ADD CONSTRAINT fk_client_init_acc_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4071 (class 2606 OID 17576)
-- Name: component_config fk_component_config; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.component_config
    ADD CONSTRAINT fk_component_config FOREIGN KEY (component_id) REFERENCES public.component(id);


--
-- TOC entry 4072 (class 2606 OID 17571)
-- Name: component fk_component_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.component
    ADD CONSTRAINT fk_component_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4056 (class 2606 OID 17276)
-- Name: realm_default_groups fk_def_groups_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_default_groups
    ADD CONSTRAINT fk_def_groups_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4049 (class 2606 OID 17136)
-- Name: user_federation_mapper_config fk_fedmapper_cfg; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_mapper_config
    ADD CONSTRAINT fk_fedmapper_cfg FOREIGN KEY (user_federation_mapper_id) REFERENCES public.user_federation_mapper(id);


--
-- TOC entry 4047 (class 2606 OID 17131)
-- Name: user_federation_mapper fk_fedmapperpm_fedprv; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_mapper
    ADD CONSTRAINT fk_fedmapperpm_fedprv FOREIGN KEY (federation_provider_id) REFERENCES public.user_federation_provider(id);


--
-- TOC entry 4048 (class 2606 OID 17126)
-- Name: user_federation_mapper fk_fedmapperpm_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_mapper
    ADD CONSTRAINT fk_fedmapperpm_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4069 (class 2606 OID 17494)
-- Name: associated_policy fk_frsr5s213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.associated_policy
    ADD CONSTRAINT fk_frsr5s213xcx4wnkog82ssrfy FOREIGN KEY (associated_policy_id) REFERENCES public.resource_server_policy(id);


--
-- TOC entry 4067 (class 2606 OID 17479)
-- Name: scope_policy fk_frsrasp13xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.scope_policy
    ADD CONSTRAINT fk_frsrasp13xcx4wnkog82ssrfy FOREIGN KEY (policy_id) REFERENCES public.resource_server_policy(id);


--
-- TOC entry 4076 (class 2606 OID 17851)
-- Name: resource_server_perm_ticket fk_frsrho213xcx4wnkog82sspmt; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_perm_ticket
    ADD CONSTRAINT fk_frsrho213xcx4wnkog82sspmt FOREIGN KEY (resource_server_id) REFERENCES public.resource_server(id);


--
-- TOC entry 4059 (class 2606 OID 17695)
-- Name: resource_server_resource fk_frsrho213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_resource
    ADD CONSTRAINT fk_frsrho213xcx4wnkog82ssrfy FOREIGN KEY (resource_server_id) REFERENCES public.resource_server(id);


--
-- TOC entry 4077 (class 2606 OID 17856)
-- Name: resource_server_perm_ticket fk_frsrho213xcx4wnkog83sspmt; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_perm_ticket
    ADD CONSTRAINT fk_frsrho213xcx4wnkog83sspmt FOREIGN KEY (resource_id) REFERENCES public.resource_server_resource(id);


--
-- TOC entry 4078 (class 2606 OID 17861)
-- Name: resource_server_perm_ticket fk_frsrho213xcx4wnkog84sspmt; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_perm_ticket
    ADD CONSTRAINT fk_frsrho213xcx4wnkog84sspmt FOREIGN KEY (scope_id) REFERENCES public.resource_server_scope(id);


--
-- TOC entry 4070 (class 2606 OID 17489)
-- Name: associated_policy fk_frsrpas14xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.associated_policy
    ADD CONSTRAINT fk_frsrpas14xcx4wnkog82ssrfy FOREIGN KEY (policy_id) REFERENCES public.resource_server_policy(id);


--
-- TOC entry 4068 (class 2606 OID 17474)
-- Name: scope_policy fk_frsrpass3xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.scope_policy
    ADD CONSTRAINT fk_frsrpass3xcx4wnkog82ssrfy FOREIGN KEY (scope_id) REFERENCES public.resource_server_scope(id);


--
-- TOC entry 4079 (class 2606 OID 17883)
-- Name: resource_server_perm_ticket fk_frsrpo2128cx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_perm_ticket
    ADD CONSTRAINT fk_frsrpo2128cx4wnkog82ssrfy FOREIGN KEY (policy_id) REFERENCES public.resource_server_policy(id);


--
-- TOC entry 4061 (class 2606 OID 17690)
-- Name: resource_server_policy fk_frsrpo213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_policy
    ADD CONSTRAINT fk_frsrpo213xcx4wnkog82ssrfy FOREIGN KEY (resource_server_id) REFERENCES public.resource_server(id);


--
-- TOC entry 4063 (class 2606 OID 17444)
-- Name: resource_scope fk_frsrpos13xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_scope
    ADD CONSTRAINT fk_frsrpos13xcx4wnkog82ssrfy FOREIGN KEY (resource_id) REFERENCES public.resource_server_resource(id);


--
-- TOC entry 4065 (class 2606 OID 17459)
-- Name: resource_policy fk_frsrpos53xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_policy
    ADD CONSTRAINT fk_frsrpos53xcx4wnkog82ssrfy FOREIGN KEY (resource_id) REFERENCES public.resource_server_resource(id);


--
-- TOC entry 4066 (class 2606 OID 17464)
-- Name: resource_policy fk_frsrpp213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_policy
    ADD CONSTRAINT fk_frsrpp213xcx4wnkog82ssrfy FOREIGN KEY (policy_id) REFERENCES public.resource_server_policy(id);


--
-- TOC entry 4064 (class 2606 OID 17449)
-- Name: resource_scope fk_frsrps213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_scope
    ADD CONSTRAINT fk_frsrps213xcx4wnkog82ssrfy FOREIGN KEY (scope_id) REFERENCES public.resource_server_scope(id);


--
-- TOC entry 4060 (class 2606 OID 17700)
-- Name: resource_server_scope fk_frsrso213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_server_scope
    ADD CONSTRAINT fk_frsrso213xcx4wnkog82ssrfy FOREIGN KEY (resource_server_id) REFERENCES public.resource_server(id);


--
-- TOC entry 4012 (class 2606 OID 16710)
-- Name: composite_role fk_gr7thllb9lu8q4vqa4524jjy8; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.composite_role
    ADD CONSTRAINT fk_gr7thllb9lu8q4vqa4524jjy8 FOREIGN KEY (child_role) REFERENCES public.keycloak_role(id);


--
-- TOC entry 4075 (class 2606 OID 17826)
-- Name: user_consent_client_scope fk_grntcsnt_clsc_usc; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_consent_client_scope
    ADD CONSTRAINT fk_grntcsnt_clsc_usc FOREIGN KEY (user_consent_id) REFERENCES public.user_consent(id);


--
-- TOC entry 4041 (class 2606 OID 16990)
-- Name: user_consent fk_grntcsnt_user; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_consent
    ADD CONSTRAINT fk_grntcsnt_user FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4054 (class 2606 OID 17250)
-- Name: group_attribute fk_group_attribute_group; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.group_attribute
    ADD CONSTRAINT fk_group_attribute_group FOREIGN KEY (group_id) REFERENCES public.keycloak_group(id);


--
-- TOC entry 4053 (class 2606 OID 17264)
-- Name: group_role_mapping fk_group_role_group; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.group_role_mapping
    ADD CONSTRAINT fk_group_role_group FOREIGN KEY (group_id) REFERENCES public.keycloak_group(id);


--
-- TOC entry 4038 (class 2606 OID 16936)
-- Name: realm_enabled_event_types fk_h846o4h0w8epx5nwedrf5y69j; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_enabled_event_types
    ADD CONSTRAINT fk_h846o4h0w8epx5nwedrf5y69j FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4016 (class 2606 OID 16720)
-- Name: realm_events_listeners fk_h846o4h0w8epx5nxev9f5y69j; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_events_listeners
    ADD CONSTRAINT fk_h846o4h0w8epx5nxev9f5y69j FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4039 (class 2606 OID 16980)
-- Name: identity_provider_mapper fk_idpm_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider_mapper
    ADD CONSTRAINT fk_idpm_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4040 (class 2606 OID 17150)
-- Name: idp_mapper_config fk_idpmconfig; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.idp_mapper_config
    ADD CONSTRAINT fk_idpmconfig FOREIGN KEY (idp_mapper_id) REFERENCES public.identity_provider_mapper(id);


--
-- TOC entry 4026 (class 2606 OID 16730)
-- Name: web_origins fk_lojpho213xcx4wnkog82ssrfy; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.web_origins
    ADD CONSTRAINT fk_lojpho213xcx4wnkog82ssrfy FOREIGN KEY (client_id) REFERENCES public.client(id);


--
-- TOC entry 4020 (class 2606 OID 16740)
-- Name: scope_mapping fk_ouse064plmlr732lxjcn1q5f1; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.scope_mapping
    ADD CONSTRAINT fk_ouse064plmlr732lxjcn1q5f1 FOREIGN KEY (client_id) REFERENCES public.client(id);


--
-- TOC entry 4031 (class 2606 OID 16875)
-- Name: protocol_mapper fk_pcm_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.protocol_mapper
    ADD CONSTRAINT fk_pcm_realm FOREIGN KEY (client_id) REFERENCES public.client(id);


--
-- TOC entry 4013 (class 2606 OID 16755)
-- Name: credential fk_pfyr0glasqyl0dei3kl69r6v0; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.credential
    ADD CONSTRAINT fk_pfyr0glasqyl0dei3kl69r6v0 FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4032 (class 2606 OID 17143)
-- Name: protocol_mapper_config fk_pmconfig; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.protocol_mapper_config
    ADD CONSTRAINT fk_pmconfig FOREIGN KEY (protocol_mapper_id) REFERENCES public.protocol_mapper(id);


--
-- TOC entry 4074 (class 2606 OID 17811)
-- Name: default_client_scope fk_r_def_cli_scope_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.default_client_scope
    ADD CONSTRAINT fk_r_def_cli_scope_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4052 (class 2606 OID 17185)
-- Name: required_action_provider fk_req_act_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.required_action_provider
    ADD CONSTRAINT fk_req_act_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4081 (class 2606 OID 17891)
-- Name: resource_uris fk_resource_server_uris; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.resource_uris
    ADD CONSTRAINT fk_resource_server_uris FOREIGN KEY (resource_id) REFERENCES public.resource_server_resource(id);


--
-- TOC entry 4082 (class 2606 OID 17905)
-- Name: role_attribute fk_role_attribute_id; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.role_attribute
    ADD CONSTRAINT fk_role_attribute_id FOREIGN KEY (role_id) REFERENCES public.keycloak_role(id);


--
-- TOC entry 4036 (class 2606 OID 16905)
-- Name: realm_supported_locales fk_supported_locales_realm; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.realm_supported_locales
    ADD CONSTRAINT fk_supported_locales_realm FOREIGN KEY (realm_id) REFERENCES public.realm(id);


--
-- TOC entry 4022 (class 2606 OID 16775)
-- Name: user_federation_config fk_t13hpu1j94r2ebpekr39x5eu5; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_federation_config
    ADD CONSTRAINT fk_t13hpu1j94r2ebpekr39x5eu5 FOREIGN KEY (user_federation_provider_id) REFERENCES public.user_federation_provider(id);


--
-- TOC entry 4055 (class 2606 OID 17257)
-- Name: user_group_membership fk_user_group_user; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.user_group_membership
    ADD CONSTRAINT fk_user_group_user FOREIGN KEY (user_id) REFERENCES public.user_entity(id);


--
-- TOC entry 4062 (class 2606 OID 17434)
-- Name: policy_config fkdc34197cf864c4e43; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.policy_config
    ADD CONSTRAINT fkdc34197cf864c4e43 FOREIGN KEY (policy_id) REFERENCES public.resource_server_policy(id);


--
-- TOC entry 4035 (class 2606 OID 16885)
-- Name: identity_provider_config fkdc4897cf864c4e43; Type: FK CONSTRAINT; Schema: public; Owner: keycloak
--

ALTER TABLE ONLY public.identity_provider_config
    ADD CONSTRAINT fkdc4897cf864c4e43 FOREIGN KEY (identity_provider_id) REFERENCES public.identity_provider(internal_id);


-- Completed on 2025-02-25 12:32:19 UTC

--
-- PostgreSQL database dump complete
--

INSERT INTO realm VALUES ('master',60,300,60,NULL,NULL,NULL,true,false,0,NULL,'master',0,NULL,false,false,false,false,'EXTERNAL',1800,36000,false,false,'e6b04c6f-e451-49ce-95b1-01b3325b77f7',1800,false,NULL,false,false,false,false,0,1,30,6,'HmacSHA1','totp','5adc5270-4510-48ee-898c-cbae8f28a3cd','a1a80a32-2bfd-4d72-9702-b41896868e69','7083d693-5892-42b4-a592-f63c604dd8dc','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6','d180151b-30b4-4438-8ec3-9033d2e60c38',2592000,false,900,true,false,'df0e3e20-eeea-4996-ba53-8ec117fcbec1',0,false,0,0,'4d227aaf-b4fa-4a86-9535-30210f612f2e');
INSERT INTO authentication_flow VALUES ('124c5389-5d37-4d77-a226-f39350c568e9','Direct Grant - Conditional OTP','Flow to determine if the OTP is required for the authentication','master','basic-flow',false,true),('12af5db4-211e-40f6-9805-89e2b48a0b50','Verify Existing Account by Re-authentication','Reauthentication of existing account','master','basic-flow',false,true),('1bd473b0-79f0-46a5-a47c-fc9fad029026','Authentication Options','Authentication options.','master','basic-flow',false,true),('2a4ea6d6-302a-44dc-8b3d-e57dad183c9f','First broker login - Conditional OTP','Flow to determine if the OTP is required for the authentication','master','basic-flow',false,true),('3dc0e0a6-fec2-47b9-b48d-e31bf31a164d','first broker login','Actions taken after first broker login with identity provider account, which is not yet linked to any Keycloak account','master','basic-flow',true,true),('3e2f3567-7f57-4dfd-95b1-90a6fff2e79c','Account verification options','Method with which to verity the existing account','master','basic-flow',false,true),('5157a4d6-247e-447e-bb54-12e9136c3dc4','User creation or linking','Flow for the existing/non-existing user alternatives','master','basic-flow',false,true),('5adc5270-4510-48ee-898c-cbae8f28a3cd','browser','browser based authentication','master','basic-flow',true,true),('5fdff352-e4e5-4dd0-8eb9-a4c8462cb539','saml ecp','SAML ECP Profile Authentication Flow','master','basic-flow',true,true),('6137a5dc-223b-4817-aed0-4b33b126164d','http challenge','An authentication flow based on challenge-response HTTP Authentication Schemes','master','basic-flow',true,true),('7083d693-5892-42b4-a592-f63c604dd8dc','direct grant','OpenID Connect Resource Owner Grant','master','basic-flow',true,true),('7d35ce2d-9375-42b1-ba30-249cdbbb1944','forms','Username, password, otp and other auth forms.','master','basic-flow',false,true),('8a75833c-68ef-4248-9602-5781e764bb55','registration form','registration form','master','form-flow',false,true),('9e586e10-7d4f-4ad6-977a-fd05dffd6ee6','reset credentials','Reset credentials for a user if they forgot their password or something','master','basic-flow',true,true),('9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a','Reset - Conditional OTP','Flow to determine if the OTP should be reset or not. Set to REQUIRED to force.','master','basic-flow',false,true),('a1a80a32-2bfd-4d72-9702-b41896868e69','registration','registration flow','master','basic-flow',true,true),('a6df1ac7-e51e-416c-afc2-3fb82e5594c3','Handle Existing Account','Handle what to do if there is existing account with same email/username like authenticated identity provider','master','basic-flow',false,true),('b9691c0e-fca4-43f2-bca1-40558f346594','Browser - Conditional OTP','Flow to determine if the OTP is required for the authentication','master','basic-flow',false,true),('d180151b-30b4-4438-8ec3-9033d2e60c38','clients','Base authentication for clients','master','client-flow',true,true),('df0e3e20-eeea-4996-ba53-8ec117fcbec1','docker auth','Used by Docker clients to authenticate against the IDP','master','basic-flow',true,true);
INSERT INTO authentication_execution VALUES ('10d29d2a-e6ef-412f-9d8d-a73245d94886',NULL,'direct-grant-validate-otp','master','124c5389-5d37-4d77-a226-f39350c568e9',0,20,false,NULL,NULL),('10ff89c6-f5b0-4076-8763-a2d9f5467c1a',NULL,'client-jwt','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,20,false,NULL,NULL),('15633fb6-e4d4-48a1-817b-dd22e177de5c',NULL,'conditional-user-configured','master','b9691c0e-fca4-43f2-bca1-40558f346594',0,10,false,NULL,NULL),('1723bba0-6f63-4c54-baff-8e96e803ea4e',NULL,'identity-provider-redirector','master','5adc5270-4510-48ee-898c-cbae8f28a3cd',2,25,false,NULL,NULL),('235d8c25-fd4c-4631-baaf-4c46259b3ed5',NULL,'auth-spnego','master','5adc5270-4510-48ee-898c-cbae8f28a3cd',3,20,false,NULL,NULL),('236ec61c-6e8d-48d3-8453-03d092f38445',NULL,'idp-email-verification','master','3e2f3567-7f57-4dfd-95b1-90a6fff2e79c',2,10,false,NULL,NULL),('27880ce3-5679-4fbe-a0ee-680ad9da4f5a',NULL,'client-x509','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,40,false,NULL,NULL),('295cf641-f50a-445c-a3cf-a43afcd5c6d6',NULL,'idp-review-profile','master','3dc0e0a6-fec2-47b9-b48d-e31bf31a164d',0,10,false,NULL,'98d35458-7329-4099-9c58-c12984adf496'),('2f647074-0625-441d-a125-ebbb93a73af6',NULL,'registration-password-action','master','8a75833c-68ef-4248-9602-5781e764bb55',0,50,false,NULL,NULL),('3034a537-c4af-44eb-8f29-6231d566f6b5',NULL,NULL,'master','5157a4d6-247e-447e-bb54-12e9136c3dc4',2,20,true,'a6df1ac7-e51e-416c-afc2-3fb82e5594c3',NULL),('38807fba-6bb3-4563-9165-5de5e02ddf8e',NULL,'conditional-user-configured','master','124c5389-5d37-4d77-a226-f39350c568e9',0,10,false,NULL,NULL),('3ad4351a-b76e-4d6f-b1ac-2a6e4e28f9e3',NULL,NULL,'master','6137a5dc-223b-4817-aed0-4b33b126164d',0,20,true,'1bd473b0-79f0-46a5-a47c-fc9fad029026',NULL),('3d96b58d-1156-48cc-9727-a038a34f3067',NULL,'registration-profile-action','master','8a75833c-68ef-4248-9602-5781e764bb55',0,40,false,NULL,NULL),('4283107d-e756-40f0-ba37-df7e311113f7',NULL,'auth-otp-form','master','b9691c0e-fca4-43f2-bca1-40558f346594',0,20,false,NULL,NULL),('43bcf0f9-6d2d-411b-8f19-28fa076b8bce',NULL,NULL,'master','5adc5270-4510-48ee-898c-cbae8f28a3cd',2,30,true,'7d35ce2d-9375-42b1-ba30-249cdbbb1944',NULL),('539c9f31-df42-4dcb-a128-9359bbb553e9',NULL,'idp-create-user-if-unique','master','5157a4d6-247e-447e-bb54-12e9136c3dc4',2,10,false,NULL,'d3806a77-7749-48bb-bbd8-0981d7bad74f'),('5f94d282-18e2-422a-86cc-3a79b79687c4',NULL,'reset-otp','master','9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a',0,20,false,NULL,NULL),('667e16c2-69d5-4f29-b37b-9502736ff929',NULL,'auth-cookie','master','5adc5270-4510-48ee-898c-cbae8f28a3cd',2,10,false,NULL,NULL),('67410c98-7dae-4c92-9c94-14614aaefe91',NULL,NULL,'master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',1,40,true,'9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a',NULL),('681e8bca-e83a-410f-8ede-bedbd406bf5f',NULL,NULL,'master','3e2f3567-7f57-4dfd-95b1-90a6fff2e79c',2,20,true,'12af5db4-211e-40f6-9805-89e2b48a0b50',NULL),('6a3101e1-a32f-4545-872c-59160b678493',NULL,'registration-user-creation','master','8a75833c-68ef-4248-9602-5781e764bb55',0,20,false,NULL,NULL),('6b937d13-a35f-4ff2-bfe5-b7efe66def55',NULL,'reset-credential-email','master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',0,20,false,NULL,NULL),('7be2fcbe-d927-4534-a87c-e83ca5e2d015',NULL,NULL,'master','7083d693-5892-42b4-a592-f63c604dd8dc',1,30,true,'124c5389-5d37-4d77-a226-f39350c568e9',NULL),('7d145ace-a68b-4b81-ae70-5c49cec28d7c',NULL,NULL,'master','3dc0e0a6-fec2-47b9-b48d-e31bf31a164d',0,20,true,'5157a4d6-247e-447e-bb54-12e9136c3dc4',NULL),('8152d141-281c-4bf2-a62c-814a89209e7d',NULL,'direct-grant-validate-username','master','7083d693-5892-42b4-a592-f63c604dd8dc',0,10,false,NULL,NULL),('87b696c3-9d25-4ba3-83a3-39b849c0629f',NULL,'basic-auth','master','1bd473b0-79f0-46a5-a47c-fc9fad029026',0,10,false,NULL,NULL),('87ea521b-38be-490a-9c91-ba9d97d39583',NULL,'auth-spnego','master','1bd473b0-79f0-46a5-a47c-fc9fad029026',3,30,false,NULL,NULL),('898c48ac-05f7-46c1-9fd1-a4a5618ddebb',NULL,'registration-page-form','master','a1a80a32-2bfd-4d72-9702-b41896868e69',0,10,true,'8a75833c-68ef-4248-9602-5781e764bb55',NULL),('8d478b03-61d0-4c84-bd80-b196e083455d',NULL,'auth-otp-form','master','2a4ea6d6-302a-44dc-8b3d-e57dad183c9f',0,20,false,NULL,NULL),('96954bdc-28b2-4234-b6e8-98b3d97c9ab8',NULL,'client-secret','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,10,false,NULL,NULL),('a0a1ccbd-2245-4ba6-b4fd-ce08c3353535',NULL,'http-basic-authenticator','master','5fdff352-e4e5-4dd0-8eb9-a4c8462cb539',0,10,false,NULL,NULL),('a1509e28-8554-4ce1-85b5-bb44d0c0a4fe',NULL,'idp-confirm-link','master','a6df1ac7-e51e-416c-afc2-3fb82e5594c3',0,10,false,NULL,NULL),('a15b00a9-e28f-41d4-89bf-32bac833e6d5',NULL,'no-cookie-redirect','master','6137a5dc-223b-4817-aed0-4b33b126164d',0,10,false,NULL,NULL),('a168dedd-ade1-437d-93f1-a56ea68e27ee',NULL,NULL,'master','a6df1ac7-e51e-416c-afc2-3fb82e5594c3',0,20,true,'3e2f3567-7f57-4dfd-95b1-90a6fff2e79c',NULL),('a4c4e437-9d6e-4639-b81d-55be2b3f0b0e',NULL,'idp-username-password-form','master','12af5db4-211e-40f6-9805-89e2b48a0b50',0,10,false,NULL,NULL),('adfc31b0-9a5a-4137-aff6-71c7800c2aed',NULL,'direct-grant-validate-password','master','7083d693-5892-42b4-a592-f63c604dd8dc',0,20,false,NULL,NULL),('b62a0854-bf37-4886-9e9f-0a4b14cdc1b1',NULL,NULL,'master','12af5db4-211e-40f6-9805-89e2b48a0b50',1,20,true,'2a4ea6d6-302a-44dc-8b3d-e57dad183c9f',NULL),('bdc22f68-2bb6-422b-ac66-d4fbb52592fb',NULL,'auth-username-password-form','master','7d35ce2d-9375-42b1-ba30-249cdbbb1944',0,10,false,NULL,NULL),('c3c7c44b-e5bf-4929-873e-7a8f17659089',NULL,NULL,'master','7d35ce2d-9375-42b1-ba30-249cdbbb1944',1,20,true,'b9691c0e-fca4-43f2-bca1-40558f346594',NULL),('c86853bf-6edb-4e93-9499-c779186f4c63',NULL,'reset-credentials-choose-user','master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',0,10,false,NULL,NULL),('d6afca0a-e5f2-4532-b99e-4e73b9eb0545',NULL,'basic-auth-otp','master','1bd473b0-79f0-46a5-a47c-fc9fad029026',3,20,false,NULL,NULL),('d9a09526-9cb4-4fea-8c2c-d6dde18cfcbe',NULL,'client-secret-jwt','master','d180151b-30b4-4438-8ec3-9033d2e60c38',2,30,false,NULL,NULL),('da6090f5-1109-4f24-a199-dec7dcdb202e',NULL,'conditional-user-configured','master','2a4ea6d6-302a-44dc-8b3d-e57dad183c9f',0,10,false,NULL,NULL),('df617d17-ac56-46cc-bb90-b70211881606',NULL,'reset-password','master','9e586e10-7d4f-4ad6-977a-fd05dffd6ee6',0,30,false,NULL,NULL),('e50e9d13-59af-4a2b-9cf6-10b64d6583de',NULL,'conditional-user-configured','master','9f0eb3c3-59a5-43ae-aa6c-fc70b158ea7a',0,10,false,NULL,NULL),('f18aa542-0611-4ec3-8045-3c2084d97c35',NULL,'docker-http-basic-authenticator','master','df0e3e20-eeea-4996-ba53-8ec117fcbec1',0,10,false,NULL,NULL),('f486afcd-90a1-469a-8355-6e10b4abe9c1',NULL,'registration-recaptcha-action','master','8a75833c-68ef-4248-9602-5781e764bb55',3,60,false,NULL,NULL);
INSERT INTO authenticator_config VALUES ('98d35458-7329-4099-9c58-c12984adf496','review profile config','master'),('d3806a77-7749-48bb-bbd8-0981d7bad74f','create unique user config','master');
INSERT INTO authenticator_config_entry VALUES ('98d35458-7329-4099-9c58-c12984adf496','missing','update.profile.on.first.login'),('d3806a77-7749-48bb-bbd8-0981d7bad74f','false','require.password.update.after.registration');
INSERT INTO client VALUES ('4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,false,'account',0,true,NULL,'/realms/master/account/',false,NULL,false,'master','openid-connect',0,false,false,'${client_account}',false,'client-secret','${authBaseUrl}',NULL,NULL,true,false,false,false),('54905dd0-4ade-494e-9c35-ab2d445a99f5',true,false,'account-console',0,true,NULL,'/realms/master/account/',false,NULL,false,'master','openid-connect',0,false,false,'${client_account-console}',false,'client-secret','${authBaseUrl}',NULL,NULL,true,false,false,false),('5a059221-51fd-434f-84a6-40fa51cda5ce',true,true,'photoprism-develop',0,false,'9d8351a0-ca01-4556-9c37-85eb634869b9',NULL,false,'https://app.localssl.dev/',false,'master','openid-connect',-1,false,false,'PhotoPrism',false,'client-secret','https://app.localssl.dev/',NULL,NULL,true,false,true,false),('5b62e4f6-f646-4e0b-aa07-83a17a324137',true,false,'broker',0,false,NULL,NULL,true,NULL,false,'master','openid-connect',0,false,false,'${client_broker}',false,'client-secret',NULL,NULL,NULL,true,false,false,false),('8a6bade2-ad19-45f1-9923-b357684d765c',true,false,'admin-cli',0,true,NULL,NULL,false,NULL,false,'master','openid-connect',0,false,false,'${client_admin-cli}',false,'client-secret',NULL,NULL,NULL,false,false,true,false),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc',true,false,'security-admin-console',0,true,NULL,'/admin/master/console/',false,NULL,false,'master','openid-connect',0,false,false,'${client_security-admin-console}',false,'client-secret','${authAdminUrl}',NULL,NULL,true,false,false,false),('e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,false,'master-realm',0,false,NULL,NULL,true,NULL,false,'master',NULL,0,false,false,'master Realm',false,'client-secret',NULL,NULL,NULL,true,false,false,false);
INSERT INTO client_attributes VALUES ('54905dd0-4ade-494e-9c35-ab2d445a99f5','pkce.code.challenge.method','S256'),('5a059221-51fd-434f-84a6-40fa51cda5ce','backchannel.logout.revoke.offline.tokens','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','backchannel.logout.session.required','true'),('5a059221-51fd-434f-84a6-40fa51cda5ce','client_credentials.use_refresh_token','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','display.on.consent.screen','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','exclude.session.state.from.auth.response','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','id.token.as.detached.signature','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','oauth2.device.authorization.grant.enabled','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','oidc.ciba.grant.enabled','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','require.pushed.authorization.requests','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml_force_name_id_format','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.artifact.binding','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.assertion.signature','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.authnstatement','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.client.signature','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.encrypt','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.force.post.binding','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.multivalued.roles','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.onetimeuse.condition','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.server.signature','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','saml.server.signature.keyinfo.ext','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','tls.client.certificate.bound.access.tokens','false'),('5a059221-51fd-434f-84a6-40fa51cda5ce','use.refresh.tokens','true'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','pkce.code.challenge.method','S256');
INSERT INTO client_scope VALUES ('13052fde-d239-4154-b80b-0f406ed76ded','phone','master','OpenID Connect built-in scope: phone','openid-connect'),('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','offline_access','master','OpenID Connect built-in scope: offline_access','openid-connect'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','profile','master','OpenID Connect built-in scope: profile','openid-connect'),('abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','microprofile-jwt','master','Microprofile - JWT built-in scope','openid-connect'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','address','master','OpenID Connect built-in scope: address','openid-connect'),('cb25d275-eff3-4655-b032-e163a0a23c0f','email','master','OpenID Connect built-in scope: email','openid-connect'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','web-origins','master','OpenID Connect scope for add allowed web origins to the access token','openid-connect'),('f0e07760-3d3d-45d5-b651-403f8b19de35','roles','master','OpenID Connect scope for add user roles to the access token','openid-connect'),('f55ceb89-6d3c-4bcb-882e-44c498d8b305','role_list','master','SAML role list','saml');
INSERT INTO client_scope_attributes VALUES ('13052fde-d239-4154-b80b-0f406ed76ded','${phoneScopeConsentText}','consent.screen.text'),('13052fde-d239-4154-b80b-0f406ed76ded','true','display.on.consent.screen'),('13052fde-d239-4154-b80b-0f406ed76ded','true','include.in.token.scope'),('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','${offlineAccessScopeConsentText}','consent.screen.text'),('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','true','display.on.consent.screen'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','${profileScopeConsentText}','consent.screen.text'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','true','display.on.consent.screen'),('a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb','true','include.in.token.scope'),('abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','false','display.on.consent.screen'),('abde17dd-48e0-4d26-a2b7-e75c04b1ac7f','true','include.in.token.scope'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','${addressScopeConsentText}','consent.screen.text'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','true','display.on.consent.screen'),('c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb','true','include.in.token.scope'),('cb25d275-eff3-4655-b032-e163a0a23c0f','${emailScopeConsentText}','consent.screen.text'),('cb25d275-eff3-4655-b032-e163a0a23c0f','true','display.on.consent.screen'),('cb25d275-eff3-4655-b032-e163a0a23c0f','true','include.in.token.scope'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','','consent.screen.text'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','false','display.on.consent.screen'),('e4f019f4-8a8a-4682-bf50-8e883c89cd03','false','include.in.token.scope'),('f0e07760-3d3d-45d5-b651-403f8b19de35','${rolesScopeConsentText}','consent.screen.text'),('f0e07760-3d3d-45d5-b651-403f8b19de35','true','display.on.consent.screen'),('f0e07760-3d3d-45d5-b651-403f8b19de35','false','include.in.token.scope'),('f55ceb89-6d3c-4bcb-882e-44c498d8b305','${samlRoleListScopeConsentText}','consent.screen.text'),('f55ceb89-6d3c-4bcb-882e-44c498d8b305','true','display.on.consent.screen');
INSERT INTO client_scope_client VALUES ('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','13052fde-d239-4154-b80b-0f406ed76ded',false),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('54905dd0-4ade-494e-9c35-ab2d445a99f5','13052fde-d239-4154-b80b-0f406ed76ded',false),('54905dd0-4ade-494e-9c35-ab2d445a99f5','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('54905dd0-4ade-494e-9c35-ab2d445a99f5','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('54905dd0-4ade-494e-9c35-ab2d445a99f5','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('54905dd0-4ade-494e-9c35-ab2d445a99f5','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('54905dd0-4ade-494e-9c35-ab2d445a99f5','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('54905dd0-4ade-494e-9c35-ab2d445a99f5','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('54905dd0-4ade-494e-9c35-ab2d445a99f5','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('5a059221-51fd-434f-84a6-40fa51cda5ce','13052fde-d239-4154-b80b-0f406ed76ded',false),('5a059221-51fd-434f-84a6-40fa51cda5ce','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('5a059221-51fd-434f-84a6-40fa51cda5ce','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('5a059221-51fd-434f-84a6-40fa51cda5ce','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('5a059221-51fd-434f-84a6-40fa51cda5ce','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('5a059221-51fd-434f-84a6-40fa51cda5ce','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('5a059221-51fd-434f-84a6-40fa51cda5ce','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('5a059221-51fd-434f-84a6-40fa51cda5ce','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('5b62e4f6-f646-4e0b-aa07-83a17a324137','13052fde-d239-4154-b80b-0f406ed76ded',false),('5b62e4f6-f646-4e0b-aa07-83a17a324137','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('5b62e4f6-f646-4e0b-aa07-83a17a324137','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('5b62e4f6-f646-4e0b-aa07-83a17a324137','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('5b62e4f6-f646-4e0b-aa07-83a17a324137','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('5b62e4f6-f646-4e0b-aa07-83a17a324137','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('5b62e4f6-f646-4e0b-aa07-83a17a324137','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('5b62e4f6-f646-4e0b-aa07-83a17a324137','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('8a6bade2-ad19-45f1-9923-b357684d765c','13052fde-d239-4154-b80b-0f406ed76ded',false),('8a6bade2-ad19-45f1-9923-b357684d765c','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('8a6bade2-ad19-45f1-9923-b357684d765c','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('8a6bade2-ad19-45f1-9923-b357684d765c','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('8a6bade2-ad19-45f1-9923-b357684d765c','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('8a6bade2-ad19-45f1-9923-b357684d765c','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('8a6bade2-ad19-45f1-9923-b357684d765c','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('8a6bade2-ad19-45f1-9923-b357684d765c','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','13052fde-d239-4154-b80b-0f406ed76ded',false),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','13052fde-d239-4154-b80b-0f406ed76ded',false),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('e6b04c6f-e451-49ce-95b1-01b3325b77f7','f0e07760-3d3d-45d5-b651-403f8b19de35',true);
INSERT INTO client_scope_role_mapping VALUES ('395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab','e06c7506-138d-4968-9186-cd958b29e577');
INSERT INTO component VALUES ('08b208d4-dd60-4eea-9ca1-9213f05b508c','Consent Required','master','consent-required','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('12fe216f-5e34-43fb-bc92-3c11a8a6abe1','Full Scope Disabled','master','scope','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('3b786f80-369b-4cc3-a54f-e7efa9dfca00','rsa-generated','master','rsa-generated','org.keycloak.keys.KeyProvider','master',NULL),('41e477a9-2312-4054-9bb2-48c803f200a5','hmac-generated','master','hmac-generated','org.keycloak.keys.KeyProvider','master',NULL),('4bfbde17-6bcf-4542-887f-8adb3ad83893','Max Clients Limit','master','max-clients','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('522398f4-90ab-4f2b-bdbf-9b8065f6533e','aes-generated','master','aes-generated','org.keycloak.keys.KeyProvider','master',NULL),('59a894bc-218d-4acf-9a44-6d79def59870','Allowed Client Scopes','master','allowed-client-templates','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('690008b2-bc5a-42f4-82f3-2ff750fe6e9a','Allowed Protocol Mapper Types','master','allowed-protocol-mappers','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','authenticated'),('984d324a-40e0-448a-bef3-11bf65cb9723','rsa-enc-generated','master','rsa-generated','org.keycloak.keys.KeyProvider','master',NULL),('99846733-ac9b-4ec5-8d38-df7b359d016f','Trusted Hosts','master','trusted-hosts','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('a8274b71-1b85-4cef-bcfc-ae84d37ab194','Allowed Protocol Mapper Types','master','allowed-protocol-mappers','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','anonymous'),('f6858356-dd21-4e56-892e-35b6f776978c','Allowed Client Scopes','master','allowed-client-templates','org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy','master','authenticated');
INSERT INTO component_config VALUES ('025dc769-74da-46ea-992e-651f8ed8b94b','984d324a-40e0-448a-bef3-11bf65cb9723','priority','100'),('050174df-d8b7-43ed-b790-0cf9ee4220d0','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-address-mapper'),('108d3e53-4f6d-4997-9e28-e6907c2220ac','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','saml-user-property-mapper'),('1245310b-9eb4-4fe6-bad2-ff739d504d58','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','saml-user-attribute-mapper'),('181e98a6-98bb-4d41-b357-7291d4d045a5','4bfbde17-6bcf-4542-887f-8adb3ad83893','max-clients','200'),('1c430fc1-0647-43ef-8db0-e0f5c4bcf771','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','saml-role-list-mapper'),('1e83b262-7ef9-458e-b231-443fd9030279','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-full-name-mapper'),('2192a5f2-2fd7-4eec-9d32-d6b2ca41c992','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-full-name-mapper'),('26841fbc-8552-454a-9534-db66a7d2f641','59a894bc-218d-4acf-9a44-6d79def59870','allow-default-scopes','true'),('2bca5418-276f-4310-8ec8-88fd49d92fdf','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','saml-role-list-mapper'),('351d0ecc-97e7-4dda-9aee-fd74bfa4914c','984d324a-40e0-448a-bef3-11bf65cb9723','privateKey','MIIEowIBAAKCAQEAh8ppU5X7uFJ0hBJC78wR6wN6s1qVxL9u1JcwwOocRPoG8Ua1Rsv9g0P9IXJMDEBdqzWY7voMD7wngJchN301tjimgYEfHx10KkX6oWkXuk30oaSTqXwjb19ps8hRvwQWJC6P1Y/SylgwaHved14p4hpLObMOoiFc2PrCH/TXP81T2BQqXSIvfJytdnO9E90ASDX0Hhv5wIr12Kuz4arh5/b78OWWJ9oWNfIkzgQoktbDdMysofij7zQgD8mYxcfQr4DvHetk+dn6x4GJWqhA1CTgHN0aF/sltm7G9bImmsMuVf9obZg77oHcFYcXhWBXbkG3baKtKYfGq6JoYA6CKwIDAQABAoIBAAQ6mIci36EA6GIIk48WQuSXyiV1x75F2/TA9KK9Z736L2cqNZEL30xMPMDi511mT8R6OdYPcXq3+F731e/9dUPEheL4m3iDmU+LuF94f2Ws8dZq4rJfjFb2mLshnPIe9XWRAae7/+uPTYqjeO0swI8rFHaqjeUctuCHBq6qGF4DQmEqAuZK0TMiZxPQlACR9G8WAmnAV66weKlhCshOkVKcSDFoH0NMMJ/KHmVzFo3drwJkKZNXbelv8EsuUnL/Pcj7WZfPcbCbR0M8TRDL7QUWX3CYKjQHgCnj6miA/7ydsUoa7BnJEzmF885ncrqfxfPAVSHuHqG9eFh7NMEG8JECgYEAu9o+TTqoM4ju6R+iww4PW4Ny7/IfyAu8qEcR22XewnyDz+qTOTePlPfEIx6al5Zymsz41pzvliPwQ0tgr49kuuM2GRVV9RzWq2JULCjrxj4tPJbeKd6m5JqlTRru+FiuLma/LXHAf6KU4cfbQQpStYRlP4XROk3F5voX+lDG00kCgYEAuQ02z+00b1o0MzDayHWw64FZ2C0lB47oOQRnx4HGJguL5/S44TaZiOWIqxFv1lqO8hWudmE5kCGUK6WaQDhOqueebMdbLODHn1qyJ7/ZGYCYX0tYRN0tGKvlCGoTjOEyfMLysIwQTXfNtGUYMQf43U0KI5WhduPFoDuRjMKuddMCgYEAqoluld3yZRajDbBSqpFRD9s9tOcyQwGku4AJjgvlNtqjL1XdYcw25R4pSVi3L3a9hBsgrHS8bKkjrXP4ymh7Ic6zhgIAjw0nNV+G2rArm0VG/AJandgr2s0p093npD2dozJTzIXAJB8M2gv92AXvICqZYBmz4CJKz22r5ur+FUECgYAH9yal2psINATNM0wnltFPwdihMohGhANA+QySjOZ/mr2h9WnD3/rJ5r90RaLfwjQm/YHt/I9iwd9D5bP3EbVpK+Eo44fsLZzKIjhK97obm+pzJ6YcCL05M6T/MLm4tbTbo/SYXt8QxphnLHbXHXW76OYH1BgIKxPFquq/+V1TGwKBgA5QZHYLMTtLDakpAZI4rMoSXawyvDEh1vuc+s9jwT1djyQVQbLdlyPW2vO0RYRVVXBi54azrFW+xLL4FavEqbT6sdrrzUl+8TLL7bJ1btIlReQfuuzv1Na/AWc1qJIkt/aWgzYeOiy4wZosnZvckU+vHZe5b74Vn0K0AdMH+tnM'),('37670004-a54c-49ad-a95a-3a93d22301d8','522398f4-90ab-4f2b-bdbf-9b8065f6533e','secret','QYrPTNemXvuh_s8Up4yu_A'),('3d2f9154-e187-40b8-86ce-15b77737c7cb','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-usermodel-property-mapper'),('3dd3b966-67b9-4c36-879b-664038038329','f6858356-dd21-4e56-892e-35b6f776978c','allow-default-scopes','true'),('4add7b8b-2a3b-4640-a632-7193c366fd57','41e477a9-2312-4054-9bb2-48c803f200a5','algorithm','HS256'),('4d551be7-5fdf-4a30-9f35-97ab28a654e6','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-usermodel-attribute-mapper'),('4e700315-b68f-45b0-a719-bd62aadcd927','99846733-ac9b-4ec5-8d38-df7b359d016f','host-sending-registration-request-must-match','true'),('540efe8b-23ad-423e-9d12-3493e593664a','984d324a-40e0-448a-bef3-11bf65cb9723','keyUse','enc'),('642ab800-084f-4e68-b8d2-c6d37a24557d','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-usermodel-attribute-mapper'),('6c387d3e-ce0b-4570-99c3-96d47cce9fef','41e477a9-2312-4054-9bb2-48c803f200a5','secret','wE0La3ld_-jLiWfKAICSC5QOrife31YSvkmWrpm_Fwom4ksV40GWKiuTr8QA_SjPvauBFBfJ2l9JFaFLgWcomw'),('903b4806-0a50-4244-b963-fe2e3838b416','522398f4-90ab-4f2b-bdbf-9b8065f6533e','kid','f59924e8-f425-4c59-8edb-a5d0155619d6'),('9ab02801-5e80-4c01-b711-24633d916450','41e477a9-2312-4054-9bb2-48c803f200a5','priority','100'),('9f41d424-3f2b-4ab0-949b-0e6de10ed573','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-sha256-pairwise-sub-mapper'),('a25bd806-bbbc-44f7-a43b-da98672c3040','984d324a-40e0-448a-bef3-11bf65cb9723','certificate','MIICmzCCAYMCBgF83P1ndTANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjExMTAxMTkzMTA3WhcNMzExMTAxMTkzMjQ3WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCHymlTlfu4UnSEEkLvzBHrA3qzWpXEv27UlzDA6hxE+gbxRrVGy/2DQ/0hckwMQF2rNZju+gwPvCeAlyE3fTW2OKaBgR8fHXQqRfqhaRe6TfShpJOpfCNvX2mzyFG/BBYkLo/Vj9LKWDBoe953XiniGks5sw6iIVzY+sIf9Nc/zVPYFCpdIi98nK12c70T3QBINfQeG/nAivXYq7PhquHn9vvw5ZYn2hY18iTOBCiS1sN0zKyh+KPvNCAPyZjFx9CvgO8d62T52frHgYlaqEDUJOAc3RoX+yW2bsb1siaawy5V/2htmDvugdwVhxeFYFduQbdtoq0ph8aromhgDoIrAgMBAAEwDQYJKoZIhvcNAQELBQADggEBABwB04rw1lSu7fpEc+/3zbQXDQSlFjn/UtwTwEitfwiKhRXC8g125wxg0CTzc642RLDIvtghifa+A4P2x/YWuyxwKq7xQG+EroHZ8Lc1gWePXqFVwoT6++146B2tvxG69o2G8xKdxGWafLXd1CFGe3FokRRMXYWTgXJWMuo/EE+3AY61ZPcK8BI0mSASjIz5J2wA0BCEbehxaJ7x8QWaGupqfvLetkwEyPT9s7GTvGKCS9tqSJQRuPveyapybuKWZ80xvrrH3vSDJGhfiplf5G4/X1Ir8FSjwpZ7kMoqz2EfWABldZXRHzdcGfy9w5OnTb/NAk6ULOMNWQCBcILetKU='),('a57bebf1-fc77-4661-9a8f-227d92cd0099','3b786f80-369b-4cc3-a54f-e7efa9dfca00','priority','100'),('a5ed2232-35fc-476a-85b0-dabc6dd4e2fc','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-usermodel-property-mapper'),('b4dbd7d0-958c-4675-a8ab-b5f7118e4a40','3b786f80-369b-4cc3-a54f-e7efa9dfca00','certificate','MIICmzCCAYMCBgF83P1lIjANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDDAZtYXN0ZXIwHhcNMjExMTAxMTkzMTA2WhcNMzExMTAxMTkzMjQ2WjARMQ8wDQYDVQQDDAZtYXN0ZXIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDJ5qTFqw8c0tPqW5foRjxXei+WKfuudGMY13s2+QpO7qCBBJZTuo9nkj1Ij1PG19Aq6IC/fVNwaf1xk8u4m8pFj1hQKWE64t9glMu3Ew1RbRyrV1RDuu3uPxSV3gUmKbQda11rpgHlSfigouJERCMU79usgZ5URsnyxWQgUxN9iif8Psu0I+DQcc7K8JZrQw+uQkCxWrGuv6O6pnwS7m/pbm3XoV6FLM7mqUxnL+81UVCJRdpCtgtQxgw5A/8UxVBHocEkkH23IFP4R2UMak5quWtgXjMIUmAiOhE0LuHjUi+NWincWJ0/DENuvRc2Ukr//U7h4FeWEhxu2EaFcwQvAgMBAAEwDQYJKoZIhvcNAQELBQADggEBALN4mBahtUhFRFSvIDqF1vhMkri60BIIu6UHy6CivBrs0pQOAOIo/G7STJrtGyDhybAMT4J3FzRXYASIDHGDLSmGqz9gmxVfMt08EEAyM+9Ep7d1FLTiC1iLHqP5BLe9WVVAIro3/w2yqExqj8XJhUa/xpQA42lIu3didUe/13SZ8oiAgLsO55tV/Uz2uGsIfQ8FeImCEkFQNeJQZspc5RT32ARBgGP8RlpgystIbiXCUpq/l90EoNLUJZJZ3iVlCl7SwPY+eqfY86IZl40xM3RmX2U9VVEh9KWhurM04M5up2MiAoleBwjlmY3HCC08aU3lQtzK1guRfDYWy/bXQFA='),('c769fd0f-bb79-4466-ae9c-7b3c403029bd','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','saml-user-property-mapper'),('d477301c-86b9-4fdb-ba3f-7874d7f3fa45','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','saml-user-attribute-mapper'),('e033eb96-3a12-4f88-9402-31aa00f6113a','3b786f80-369b-4cc3-a54f-e7efa9dfca00','keyUse','sig'),('e24bbbfa-620b-44cb-b070-e0c2383555f2','690008b2-bc5a-42f4-82f3-2ff750fe6e9a','allowed-protocol-mapper-types','oidc-sha256-pairwise-sub-mapper'),('e3eb40d3-45f7-4edf-9347-d41334635763','41e477a9-2312-4054-9bb2-48c803f200a5','kid','7ca5813e-1813-49c6-a1f8-d1ecf58f56e9'),('eef9b1d5-555e-4b75-9e37-17e4d4b91403','522398f4-90ab-4f2b-bdbf-9b8065f6533e','priority','100'),('f5a7c7b4-ce7e-458d-839f-78e660238f6b','a8274b71-1b85-4cef-bcfc-ae84d37ab194','allowed-protocol-mapper-types','oidc-address-mapper'),('f5b7f63b-8d3a-4575-a3e6-4346cda1a549','3b786f80-369b-4cc3-a54f-e7efa9dfca00','privateKey','MIIEpQIBAAKCAQEAyeakxasPHNLT6luX6EY8V3ovlin7rnRjGNd7NvkKTu6ggQSWU7qPZ5I9SI9TxtfQKuiAv31TcGn9cZPLuJvKRY9YUClhOuLfYJTLtxMNUW0cq1dUQ7rt7j8Uld4FJim0HWtda6YB5Un4oKLiREQjFO/brIGeVEbJ8sVkIFMTfYon/D7LtCPg0HHOyvCWa0MPrkJAsVqxrr+juqZ8Eu5v6W5t16FehSzO5qlMZy/vNVFQiUXaQrYLUMYMOQP/FMVQR6HBJJB9tyBT+EdlDGpOarlrYF4zCFJgIjoRNC7h41IvjVop3FidPwxDbr0XNlJK//1O4eBXlhIcbthGhXMELwIDAQABAoIBAQCea3I4g5tNE4QiLJJKN+oa/Y2fNvv7i+lB0boljU1wV77q3Q2TTxw8uTuK1qN2r1nwgRScrBqvZwrtdnlwNhWFdQ9nfsCC8wcxAi/CS5m0nXfUXaaJqoAM48QkP9wscKaaOudHky+DmQIUERqXVBtuzzG/7sir+gt1iTqiPm1Zn4v/U7l1eRhIPx/qJ8+4iTw6crRLooUzSHqCirZWs6pRbzAoxd3OBdxr+wXqmzV1+F6HguonhawFnqilaHEf0Lj+N0Lc1vOdmzNSlJ3azV/kRvMBqsDRKPcNIfCnjeNA+gHJd4u++OzX3IDRsJtlyj51+TmYkmMnLOfsWTRHmtfxAoGBAOrZtHJ3ronteZG55DyHN3idDLW8pi783oKSB7HYNJnIXoAupdjsSyLcN3A8EhTRbvAiA0XjfeqbpT8Xg3OLOxEZi1PgndguizutkPOIDeL0Z+p0H4CHdY3p2VWSUTquojlwaOgf7iUvPWdlDD/DM1glfvH05jThg59zhlt0qULHAoGBANwVUH3cVEtnp50oG/LlzECM9eLmsSu/4APyIGLXtk/LMtu2pJPhatKMgtOfXBTrWud0q2E6PwPaDIX15EITcK8gtGIpzDOoHvb6YLwLmxaeUy4M+oHBkR1m/CMb77r5sq5DfOR9UoTr+3tVRPxlk6ES1vqrbsJNZ9N70xzFIctZAoGBAJrsrNYKT7CbYOQaLg8j4BsH91d4MGS02ZBnBv5yMxjzjiufGjcEgfhoL4YxingDRNzSgzg6f1kh/hulxkiVo4x/PmNBvL7czWq77/BHY2nBcz++BP4D3i+VAZMqp70/cLLVjc77KV2MUUSA61iwy5EtgxXYSXi+/9ZTHmH8jqAHAoGBALaGkuAfaGW1TOThC/TyMujiP1d0fkHLe22qVMPFJXWeD8r6+hmPXTnLwQDj7MmIvDazoyMa3IJESBid6zYFy3HjDNdQ1QOOjkfFNY8fjPtASboql2Qf9ktNSxWPKM6IInG2lREnAtYspMAP4wv07nArINJ6dXx+F/rkeh0lPTbZAoGAN11eATq+0KGYoF7lEV72d74roNrg+uKffn8GsP6KlmV8nqubg/QOug7c/tuEa2hqpcA5dYTFNjKqaSluKOfje+U9RTM9/wICmQYn/VBBRzG6l0oO3X5TTefy4Qtd292N1R/px/0wmF6RhlDqIW5H7NNIpki9uwG78IA5jvWknSI='),('f7f2cc71-2362-4e27-94a3-2624238eeb28','99846733-ac9b-4ec5-8d38-df7b359d016f','client-uris-must-match','true');
INSERT INTO keycloak_role VALUES ('03484350-2ad2-4c70-bc2a-f04dd90d8d44','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_view-applications}','view-applications','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('0b1ad207-9fa7-4ed3-84ba-f800607e7d09','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_view-realm}','view-realm','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('16801c90-306a-4dd9-8f04-6a79dc56ce7a','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_create-client}','create-client','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('1aa723ff-209c-4637-b3c8-8159d72e9b09','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_manage-account}','manage-account','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('1b33555e-7a77-48cd-8c2a-5b7fae1318df','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_impersonation}','impersonation','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('1d4a0b4b-1f99-4a6e-a9e1-463eec8147fd','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_view-profile}','view-profile','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('26c52158-724b-4001-b2ee-e2eb3af6e34d','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_delete-account}','delete-account','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('2a2bb3c8-e26f-4adb-9a9e-bbfb1685745e','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_manage-account-links}','manage-account-links','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('3b8b7ca4-89c8-4038-a996-c7ce380efa2c','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_view-events}','view-events','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('4d227aaf-b4fa-4a86-9535-30210f612f2e','master',false,'${role_default-roles}','default-roles-master','master',NULL,NULL),('5827ab16-b5bc-4738-b05e-89406e065439','master',false,'${role_admin}','admin','master',NULL,NULL),('70e66b49-b385-466c-95a8-7bff51eee65f','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_query-groups}','query-groups','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('766d6f0c-d33b-480c-b193-98f7dd41b5e6','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_manage-realm}','manage-realm','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('7b199ec2-c91d-4300-a565-f21a8c16b647','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_manage-authorization}','manage-authorization','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('8340533b-d129-4b11-ae94-708efcb2a14a','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_query-clients}','query-clients','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('839188a4-115b-4eab-9fa5-9573ee95fbf0','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_view-users}','view-users','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('8462f2ed-3177-493c-88af-ee6dd888b246','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_view-consent}','view-consent','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('9066e812-fa7d-4d41-8b29-7c350d08bedf','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_manage-events}','manage-events','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('982e7b3f-64e8-4d41-9323-7d308fe31b8f','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_query-realms}','query-realms','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('a433b041-695e-4da9-a275-1b6abf9184c9','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_query-users}','query-users','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('b3794300-e7e9-4f3d-af84-7d032a01df6b','master',false,'${role_uma_authorization}','uma_authorization','master',NULL,NULL),('b52948b8-e6cc-45df-aeeb-bfa8210e9378','master',false,'${role_create-realm}','create-realm','master',NULL,NULL),('c9a3a64a-f4f7-46a4-aaf0-b1af52473ce8','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_manage-clients}','manage-clients','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('ce72ef0f-db0f-4eb7-bf60-642c203cc1e9','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_view-clients}','view-clients','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('ceab846f-983c-4c02-a89e-bfac58410427','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_manage-users}','manage-users','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('d3540296-65b3-4fa5-a007-fe0feb99d6d1','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_view-identity-providers}','view-identity-providers','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('d726a590-eea5-4040-85aa-b44f10f3bf58','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_manage-identity-providers}','manage-identity-providers','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('e06c7506-138d-4968-9186-cd958b29e577','master',false,'${role_offline-access}','offline_access','master',NULL,NULL),('e498256c-c0b0-4bfa-a7c3-f10f05de507e','e6b04c6f-e451-49ce-95b1-01b3325b77f7',true,'${role_view-authorization}','view-authorization','master','e6b04c6f-e451-49ce-95b1-01b3325b77f7',NULL),('f660cd73-8ffe-4077-b73a-b26c6a24f149','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',true,'${role_manage-consent}','manage-consent','master','4e4977d6-eaa9-4245-ae4c-04d20f5436d9',NULL),('f876756a-cf00-444a-89e8-ea745420c10b','5b62e4f6-f646-4e0b-aa07-83a17a324137',true,'${role_read-token}','read-token','master','5b62e4f6-f646-4e0b-aa07-83a17a324137',NULL);
INSERT INTO composite_role VALUES ('1aa723ff-209c-4637-b3c8-8159d72e9b09','2a2bb3c8-e26f-4adb-9a9e-bbfb1685745e'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','1aa723ff-209c-4637-b3c8-8159d72e9b09'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','1d4a0b4b-1f99-4a6e-a9e1-463eec8147fd'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','b3794300-e7e9-4f3d-af84-7d032a01df6b'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','e06c7506-138d-4968-9186-cd958b29e577'),('5827ab16-b5bc-4738-b05e-89406e065439','0b1ad207-9fa7-4ed3-84ba-f800607e7d09'),('5827ab16-b5bc-4738-b05e-89406e065439','16801c90-306a-4dd9-8f04-6a79dc56ce7a'),('5827ab16-b5bc-4738-b05e-89406e065439','1b33555e-7a77-48cd-8c2a-5b7fae1318df'),('5827ab16-b5bc-4738-b05e-89406e065439','3b8b7ca4-89c8-4038-a996-c7ce380efa2c'),('5827ab16-b5bc-4738-b05e-89406e065439','70e66b49-b385-466c-95a8-7bff51eee65f'),('5827ab16-b5bc-4738-b05e-89406e065439','766d6f0c-d33b-480c-b193-98f7dd41b5e6'),('5827ab16-b5bc-4738-b05e-89406e065439','7b199ec2-c91d-4300-a565-f21a8c16b647'),('5827ab16-b5bc-4738-b05e-89406e065439','8340533b-d129-4b11-ae94-708efcb2a14a'),('5827ab16-b5bc-4738-b05e-89406e065439','839188a4-115b-4eab-9fa5-9573ee95fbf0'),('5827ab16-b5bc-4738-b05e-89406e065439','9066e812-fa7d-4d41-8b29-7c350d08bedf'),('5827ab16-b5bc-4738-b05e-89406e065439','982e7b3f-64e8-4d41-9323-7d308fe31b8f'),('5827ab16-b5bc-4738-b05e-89406e065439','a433b041-695e-4da9-a275-1b6abf9184c9'),('5827ab16-b5bc-4738-b05e-89406e065439','b52948b8-e6cc-45df-aeeb-bfa8210e9378'),('5827ab16-b5bc-4738-b05e-89406e065439','c9a3a64a-f4f7-46a4-aaf0-b1af52473ce8'),('5827ab16-b5bc-4738-b05e-89406e065439','ce72ef0f-db0f-4eb7-bf60-642c203cc1e9'),('5827ab16-b5bc-4738-b05e-89406e065439','ceab846f-983c-4c02-a89e-bfac58410427'),('5827ab16-b5bc-4738-b05e-89406e065439','d3540296-65b3-4fa5-a007-fe0feb99d6d1'),('5827ab16-b5bc-4738-b05e-89406e065439','d726a590-eea5-4040-85aa-b44f10f3bf58'),('5827ab16-b5bc-4738-b05e-89406e065439','e498256c-c0b0-4bfa-a7c3-f10f05de507e'),('839188a4-115b-4eab-9fa5-9573ee95fbf0','70e66b49-b385-466c-95a8-7bff51eee65f'),('839188a4-115b-4eab-9fa5-9573ee95fbf0','a433b041-695e-4da9-a275-1b6abf9184c9'),('ce72ef0f-db0f-4eb7-bf60-642c203cc1e9','8340533b-d129-4b11-ae94-708efcb2a14a'),('f660cd73-8ffe-4077-b73a-b26c6a24f149','8462f2ed-3177-493c-88af-ee6dd888b246');
INSERT INTO user_entity VALUES ('563bb06b-d712-48c1-9381-cd6473e18590',NULL,'87ff588a-ad31-4774-b5e2-988648080774',false,true,NULL,NULL,NULL,'master','admin',1635795167832,NULL,0),('744b396e-3cf9-4e9f-9493-61b7b188fb10','jane.doe@example.com','jane.doe@example.com',true,true,NULL,'Jane','Doe','master','user',1635845837109,NULL,0);
INSERT INTO credential VALUES ('9e4eb727-5eef-4686-b4c7-4d84282fade1',NULL,'password','563bb06b-d712-48c1-9381-cd6473e18590',1635795167967,NULL,'{\"value\":\"PEMYsigNJ+5xMOOdQkhjMh/7x9e2qKC+Mv9usfICUOwXv79Fn9Dar3fee5FJCw86tpQLP+hz2Of1m+pAksYjdg==\",\"salt\":\"IbVs55tBY3rj4s9WcG5QMg==\",\"additionalParameters\":{}}','{\"hashIterations\":27500,\"algorithm\":\"pbkdf2-sha256\",\"additionalParameters\":{}}',10),('df81f1dd-0bd8-4664-aa2b-83304f8f54c2',NULL,'password','744b396e-3cf9-4e9f-9493-61b7b188fb10',1635845862853,NULL,'{\"value\":\"SsI8poDKv1KUYXywcU8od170skwnBD3ibXh58RjvF5zzkw/Kndk7108OjomGfORghJC5Y7rnW9HI2aPhO6DdJg==\",\"salt\":\"+QSBSxjvcC7J2FwVtbH5WQ==\",\"additionalParameters\":{}}','{\"hashIterations\":27500,\"algorithm\":\"pbkdf2-sha256\",\"additionalParameters\":{}}',10);
-- INSERT INTO databasechangelog VALUES ('1.0.0.Final-KEYCLOAK-5461','sthorger@redhat.com','META-INF/jpa-changelog-1.0.0.Final.xml','2021-11-01 19:32:34',1,'EXECUTED','7:4e70412f24a3f382c82183742ec79317','createTable tableName=APPLICATION_DEFAULT_ROLES; createTable tableName=CLIENT; createTable tableName=CLIENT_SESSION; createTable tableName=CLIENT_SESSION_ROLE; createTable tableName=COMPOSITE_ROLE; createTable tableName=CREDENTIAL; createTable tab...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.0.0.Final-KEYCLOAK-5461','sthorger@redhat.com','META-INF/db2-jpa-changelog-1.0.0.Final.xml','2021-11-01 19:32:34',2,'MARK_RAN','7:cb16724583e9675711801c6875114f28','createTable tableName=APPLICATION_DEFAULT_ROLES; createTable tableName=CLIENT; createTable tableName=CLIENT_SESSION; createTable tableName=CLIENT_SESSION_ROLE; createTable tableName=COMPOSITE_ROLE; createTable tableName=CREDENTIAL; createTable tab...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.1.0.Beta1','sthorger@redhat.com','META-INF/jpa-changelog-1.1.0.Beta1.xml','2021-11-01 19:32:34',3,'EXECUTED','7:0310eb8ba07cec616460794d42ade0fa','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=CLIENT_ATTRIBUTES; createTable tableName=CLIENT_SESSION_NOTE; createTable tableName=APP_NODE_REGISTRATIONS; addColumn table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.1.0.Final','sthorger@redhat.com','META-INF/jpa-changelog-1.1.0.Final.xml','2021-11-01 19:32:34',4,'EXECUTED','7:5d25857e708c3233ef4439df1f93f012','renameColumn newColumnName=EVENT_TIME, oldColumnName=TIME, tableName=EVENT_ENTITY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.Beta1','psilva@redhat.com','META-INF/jpa-changelog-1.2.0.Beta1.xml','2021-11-01 19:32:34',5,'EXECUTED','7:c7a54a1041d58eb3817a4a883b4d4e84','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=PROTOCOL_MAPPER; createTable tableName=PROTOCOL_MAPPER_CONFIG; createTable tableName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.Beta1','psilva@redhat.com','META-INF/db2-jpa-changelog-1.2.0.Beta1.xml','2021-11-01 19:32:34',6,'MARK_RAN','7:2e01012df20974c1c2a605ef8afe25b7','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=PROTOCOL_MAPPER; createTable tableName=PROTOCOL_MAPPER_CONFIG; createTable tableName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.RC1','bburke@redhat.com','META-INF/jpa-changelog-1.2.0.CR1.xml','2021-11-01 19:32:35',7,'EXECUTED','7:0f08df48468428e0f30ee59a8ec01a41','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=MIGRATION_MODEL; createTable tableName=IDENTITY_P...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.RC1','bburke@redhat.com','META-INF/db2-jpa-changelog-1.2.0.CR1.xml','2021-11-01 19:32:35',8,'MARK_RAN','7:a77ea2ad226b345e7d689d366f185c8c','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=MIGRATION_MODEL; createTable tableName=IDENTITY_P...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.2.0.Final','keycloak','META-INF/jpa-changelog-1.2.0.Final.xml','2021-11-01 19:32:35',9,'EXECUTED','7:a3377a2059aefbf3b90ebb4c4cc8e2ab','update tableName=CLIENT; update tableName=CLIENT; update tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.3.0','bburke@redhat.com','META-INF/jpa-changelog-1.3.0.xml','2021-11-01 19:32:35',10,'EXECUTED','7:04c1dbedc2aa3e9756d1a1668e003451','delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=ADMI...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.4.0','bburke@redhat.com','META-INF/jpa-changelog-1.4.0.xml','2021-11-01 19:32:35',11,'EXECUTED','7:36ef39ed560ad07062d956db861042ba','delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.4.0','bburke@redhat.com','META-INF/db2-jpa-changelog-1.4.0.xml','2021-11-01 19:32:35',12,'MARK_RAN','7:d909180b2530479a716d3f9c9eaea3d7','delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.5.0','bburke@redhat.com','META-INF/jpa-changelog-1.5.0.xml','2021-11-01 19:32:35',13,'EXECUTED','7:cf12b04b79bea5152f165eb41f3955f6','delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1_from15','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',14,'EXECUTED','7:7e32c8f05c755e8675764e7d5f514509','addColumn tableName=REALM; addColumn tableName=KEYCLOAK_ROLE; addColumn tableName=CLIENT; createTable tableName=OFFLINE_USER_SESSION; createTable tableName=OFFLINE_CLIENT_SESSION; addPrimaryKey constraintName=CONSTRAINT_OFFL_US_SES_PK2, tableName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1_from16-pre','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',15,'MARK_RAN','7:980ba23cc0ec39cab731ce903dd01291','delete tableName=OFFLINE_CLIENT_SESSION; delete tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1_from16','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',16,'MARK_RAN','7:2fa220758991285312eb84f3b4ff5336','dropPrimaryKey constraintName=CONSTRAINT_OFFLINE_US_SES_PK, tableName=OFFLINE_USER_SESSION; dropPrimaryKey constraintName=CONSTRAINT_OFFLINE_CL_SES_PK, tableName=OFFLINE_CLIENT_SESSION; addColumn tableName=OFFLINE_USER_SESSION; update tableName=OF...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.6.1','mposolda@redhat.com','META-INF/jpa-changelog-1.6.1.xml','2021-11-01 19:32:35',17,'EXECUTED','7:d41d8cd98f00b204e9800998ecf8427e','empty','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.7.0','bburke@redhat.com','META-INF/jpa-changelog-1.7.0.xml','2021-11-01 19:32:35',18,'EXECUTED','7:91ace540896df890cc00a0490ee52bbc','createTable tableName=KEYCLOAK_GROUP; createTable tableName=GROUP_ROLE_MAPPING; createTable tableName=GROUP_ATTRIBUTE; createTable tableName=USER_GROUP_MEMBERSHIP; createTable tableName=REALM_DEFAULT_GROUPS; addColumn tableName=IDENTITY_PROVIDER; ...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0','mposolda@redhat.com','META-INF/jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',19,'EXECUTED','7:c31d1646dfa2618a9335c00e07f89f24','addColumn tableName=IDENTITY_PROVIDER; createTable tableName=CLIENT_TEMPLATE; createTable tableName=CLIENT_TEMPLATE_ATTRIBUTES; createTable tableName=TEMPLATE_SCOPE_MAPPING; dropNotNullConstraint columnName=CLIENT_ID, tableName=PROTOCOL_MAPPER; ad...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0-2','keycloak','META-INF/jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',20,'EXECUTED','7:df8bc21027a4f7cbbb01f6344e89ce07','dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; update tableName=CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0','mposolda@redhat.com','META-INF/db2-jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',21,'MARK_RAN','7:f987971fe6b37d963bc95fee2b27f8df','addColumn tableName=IDENTITY_PROVIDER; createTable tableName=CLIENT_TEMPLATE; createTable tableName=CLIENT_TEMPLATE_ATTRIBUTES; createTable tableName=TEMPLATE_SCOPE_MAPPING; dropNotNullConstraint columnName=CLIENT_ID, tableName=PROTOCOL_MAPPER; ad...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.8.0-2','keycloak','META-INF/db2-jpa-changelog-1.8.0.xml','2021-11-01 19:32:36',22,'MARK_RAN','7:df8bc21027a4f7cbbb01f6344e89ce07','dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; update tableName=CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.0','mposolda@redhat.com','META-INF/jpa-changelog-1.9.0.xml','2021-11-01 19:32:36',23,'EXECUTED','7:ed2dc7f799d19ac452cbcda56c929e47','update tableName=REALM; update tableName=REALM; update tableName=REALM; update tableName=REALM; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=REALM; update tableName=REALM; customChange; dr...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.1','keycloak','META-INF/jpa-changelog-1.9.1.xml','2021-11-01 19:32:36',24,'EXECUTED','7:80b5db88a5dda36ece5f235be8757615','modifyDataType columnName=PRIVATE_KEY, tableName=REALM; modifyDataType columnName=PUBLIC_KEY, tableName=REALM; modifyDataType columnName=CERTIFICATE, tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.1','keycloak','META-INF/db2-jpa-changelog-1.9.1.xml','2021-11-01 19:32:36',25,'MARK_RAN','7:1437310ed1305a9b93f8848f301726ce','modifyDataType columnName=PRIVATE_KEY, tableName=REALM; modifyDataType columnName=CERTIFICATE, tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('1.9.2','keycloak','META-INF/jpa-changelog-1.9.2.xml','2021-11-01 19:32:36',26,'EXECUTED','7:b82ffb34850fa0836be16deefc6a87c4','createIndex indexName=IDX_USER_EMAIL, tableName=USER_ENTITY; createIndex indexName=IDX_USER_ROLE_MAPPING, tableName=USER_ROLE_MAPPING; createIndex indexName=IDX_USER_GROUP_MAPPING, tableName=USER_GROUP_MEMBERSHIP; createIndex indexName=IDX_USER_CO...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-2.0.0','psilva@redhat.com','META-INF/jpa-changelog-authz-2.0.0.xml','2021-11-01 19:32:36',27,'EXECUTED','7:9cc98082921330d8d9266decdd4bd658','createTable tableName=RESOURCE_SERVER; addPrimaryKey constraintName=CONSTRAINT_FARS, tableName=RESOURCE_SERVER; addUniqueConstraint constraintName=UK_AU8TT6T700S9V50BU18WS5HA6, tableName=RESOURCE_SERVER; createTable tableName=RESOURCE_SERVER_RESOU...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-2.5.1','psilva@redhat.com','META-INF/jpa-changelog-authz-2.5.1.xml','2021-11-01 19:32:36',28,'EXECUTED','7:03d64aeed9cb52b969bd30a7ac0db57e','update tableName=RESOURCE_SERVER_POLICY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.1.0-KEYCLOAK-5461','bburke@redhat.com','META-INF/jpa-changelog-2.1.0.xml','2021-11-01 19:32:36',29,'EXECUTED','7:f1f9fd8710399d725b780f463c6b21cd','createTable tableName=BROKER_LINK; createTable tableName=FED_USER_ATTRIBUTE; createTable tableName=FED_USER_CONSENT; createTable tableName=FED_USER_CONSENT_ROLE; createTable tableName=FED_USER_CONSENT_PROT_MAPPER; createTable tableName=FED_USER_CR...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.2.0','bburke@redhat.com','META-INF/jpa-changelog-2.2.0.xml','2021-11-01 19:32:36',30,'EXECUTED','7:53188c3eb1107546e6f765835705b6c1','addColumn tableName=ADMIN_EVENT_ENTITY; createTable tableName=CREDENTIAL_ATTRIBUTE; createTable tableName=FED_CREDENTIAL_ATTRIBUTE; modifyDataType columnName=VALUE, tableName=CREDENTIAL; addForeignKeyConstraint baseTableName=FED_CREDENTIAL_ATTRIBU...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.3.0','bburke@redhat.com','META-INF/jpa-changelog-2.3.0.xml','2021-11-01 19:32:37',31,'EXECUTED','7:d6e6f3bc57a0c5586737d1351725d4d4','createTable tableName=FEDERATED_USER; addPrimaryKey constraintName=CONSTR_FEDERATED_USER, tableName=FEDERATED_USER; dropDefaultValue columnName=TOTP, tableName=USER_ENTITY; dropColumn columnName=TOTP, tableName=USER_ENTITY; addColumn tableName=IDE...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.4.0','bburke@redhat.com','META-INF/jpa-changelog-2.4.0.xml','2021-11-01 19:32:37',32,'EXECUTED','7:454d604fbd755d9df3fd9c6329043aa5','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0','bburke@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',33,'EXECUTED','7:57e98a3077e29caf562f7dbf80c72600','customChange; modifyDataType columnName=USER_ID, tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-unicode-oracle','hmlnarik@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',34,'MARK_RAN','7:e4c7e8f2256210aee71ddc42f538b57a','modifyDataType columnName=DESCRIPTION, tableName=AUTHENTICATION_FLOW; modifyDataType columnName=DESCRIPTION, tableName=CLIENT_TEMPLATE; modifyDataType columnName=DESCRIPTION, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=DESCRIPTION,...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-unicode-other-dbs','hmlnarik@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',35,'EXECUTED','7:09a43c97e49bc626460480aa1379b522','modifyDataType columnName=DESCRIPTION, tableName=AUTHENTICATION_FLOW; modifyDataType columnName=DESCRIPTION, tableName=CLIENT_TEMPLATE; modifyDataType columnName=DESCRIPTION, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=DESCRIPTION,...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-duplicate-email-support','slawomir@dabek.name','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',36,'EXECUTED','7:26bfc7c74fefa9126f2ce702fb775553','addColumn tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.0-unique-group-names','hmlnarik@redhat.com','META-INF/jpa-changelog-2.5.0.xml','2021-11-01 19:32:37',37,'EXECUTED','7:a161e2ae671a9020fff61e996a207377','addUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('2.5.1','bburke@redhat.com','META-INF/jpa-changelog-2.5.1.xml','2021-11-01 19:32:37',38,'EXECUTED','7:37fc1781855ac5388c494f1442b3f717','addColumn tableName=FED_USER_CONSENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.0.0','bburke@redhat.com','META-INF/jpa-changelog-3.0.0.xml','2021-11-01 19:32:37',39,'EXECUTED','7:13a27db0dae6049541136adad7261d27','addColumn tableName=IDENTITY_PROVIDER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fix','keycloak','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',40,'MARK_RAN','7:550300617e3b59e8af3a6294df8248a3','addNotNullConstraint columnName=REALM_ID, tableName=CLIENT_INITIAL_ACCESS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fix-with-keycloak-5416','keycloak','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',41,'MARK_RAN','7:e3a9482b8931481dc2772a5c07c44f17','dropIndex indexName=IDX_CLIENT_INIT_ACC_REALM, tableName=CLIENT_INITIAL_ACCESS; addNotNullConstraint columnName=REALM_ID, tableName=CLIENT_INITIAL_ACCESS; createIndex indexName=IDX_CLIENT_INIT_ACC_REALM, tableName=CLIENT_INITIAL_ACCESS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fix-offline-sessions','hmlnarik','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',42,'EXECUTED','7:72b07d85a2677cb257edb02b408f332d','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.2.0-fixed','keycloak','META-INF/jpa-changelog-3.2.0.xml','2021-11-01 19:32:37',43,'EXECUTED','7:a72a7858967bd414835d19e04d880312','addColumn tableName=REALM; dropPrimaryKey constraintName=CONSTRAINT_OFFL_CL_SES_PK2, tableName=OFFLINE_CLIENT_SESSION; dropColumn columnName=CLIENT_SESSION_ID, tableName=OFFLINE_CLIENT_SESSION; addPrimaryKey constraintName=CONSTRAINT_OFFL_CL_SES_P...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.3.0','keycloak','META-INF/jpa-changelog-3.3.0.xml','2021-11-01 19:32:37',44,'EXECUTED','7:94edff7cf9ce179e7e85f0cd78a3cf2c','addColumn tableName=USER_ENTITY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part1','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:37',45,'EXECUTED','7:6a48ce645a3525488a90fbf76adf3bb3','addColumn tableName=RESOURCE_SERVER_POLICY; addColumn tableName=RESOURCE_SERVER_RESOURCE; addColumn tableName=RESOURCE_SERVER_SCOPE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part2-KEYCLOAK-6095','hmlnarik@redhat.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:37',46,'EXECUTED','7:e64b5dcea7db06077c6e57d3b9e5ca14','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part3-fixed','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:37',47,'MARK_RAN','7:fd8cf02498f8b1e72496a20afc75178c','dropIndex indexName=IDX_RES_SERV_POL_RES_SERV, tableName=RESOURCE_SERVER_POLICY; dropIndex indexName=IDX_RES_SRV_RES_RES_SRV, tableName=RESOURCE_SERVER_RESOURCE; dropIndex indexName=IDX_RES_SRV_SCOPE_RES_SRV, tableName=RESOURCE_SERVER_SCOPE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-3.4.0.CR1-resource-server-pk-change-part3-fixed-nodropindex','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:38',48,'EXECUTED','7:542794f25aa2b1fbabb7e577d6646319','addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, tableName=RESOURCE_SERVER_POLICY; addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, tableName=RESOURCE_SERVER_RESOURCE; addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, ...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authn-3.4.0.CR1-refresh-token-max-reuse','glavoie@gmail.com','META-INF/jpa-changelog-authz-3.4.0.CR1.xml','2021-11-01 19:32:38',49,'EXECUTED','7:edad604c882df12f74941dac3cc6d650','addColumn tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.0','keycloak','META-INF/jpa-changelog-3.4.0.xml','2021-11-01 19:32:38',50,'EXECUTED','7:0f88b78b7b46480eb92690cbf5e44900','addPrimaryKey constraintName=CONSTRAINT_REALM_DEFAULT_ROLES, tableName=REALM_DEFAULT_ROLES; addPrimaryKey constraintName=CONSTRAINT_COMPOSITE_ROLE, tableName=COMPOSITE_ROLE; addPrimaryKey constraintName=CONSTR_REALM_DEFAULT_GROUPS, tableName=REALM...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.0-KEYCLOAK-5230','hmlnarik@redhat.com','META-INF/jpa-changelog-3.4.0.xml','2021-11-01 19:32:38',51,'EXECUTED','7:d560e43982611d936457c327f872dd59','createIndex indexName=IDX_FU_ATTRIBUTE, tableName=FED_USER_ATTRIBUTE; createIndex indexName=IDX_FU_CONSENT, tableName=FED_USER_CONSENT; createIndex indexName=IDX_FU_CONSENT_RU, tableName=FED_USER_CONSENT; createIndex indexName=IDX_FU_CREDENTIAL, t...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.1','psilva@redhat.com','META-INF/jpa-changelog-3.4.1.xml','2021-11-01 19:32:38',52,'EXECUTED','7:c155566c42b4d14ef07059ec3b3bbd8e','modifyDataType columnName=VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.2','keycloak','META-INF/jpa-changelog-3.4.2.xml','2021-11-01 19:32:38',53,'EXECUTED','7:b40376581f12d70f3c89ba8ddf5b7dea','update tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('3.4.2-KEYCLOAK-5172','mkanis@redhat.com','META-INF/jpa-changelog-3.4.2.xml','2021-11-01 19:32:38',54,'EXECUTED','7:a1132cc395f7b95b3646146c2e38f168','update tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-KEYCLOAK-6335','bburke@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:38',55,'EXECUTED','7:d8dc5d89c789105cfa7ca0e82cba60af','createTable tableName=CLIENT_AUTH_FLOW_BINDINGS; addPrimaryKey constraintName=C_CLI_FLOW_BIND, tableName=CLIENT_AUTH_FLOW_BINDINGS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-CLEANUP-UNUSED-TABLE','bburke@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:38',56,'EXECUTED','7:7822e0165097182e8f653c35517656a3','dropTable tableName=CLIENT_IDENTITY_PROV_MAPPING','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-KEYCLOAK-6228','bburke@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:38',57,'EXECUTED','7:c6538c29b9c9a08f9e9ea2de5c2b6375','dropUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHOGM8UEWRT, tableName=USER_CONSENT; dropNotNullConstraint columnName=CLIENT_ID, tableName=USER_CONSENT; addColumn tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHO...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.0.0-KEYCLOAK-5579-fixed','mposolda@redhat.com','META-INF/jpa-changelog-4.0.0.xml','2021-11-01 19:32:39',58,'EXECUTED','7:6d4893e36de22369cf73bcb051ded875','dropForeignKeyConstraint baseTableName=CLIENT_TEMPLATE_ATTRIBUTES, constraintName=FK_CL_TEMPL_ATTR_TEMPL; renameTable newTableName=CLIENT_SCOPE_ATTRIBUTES, oldTableName=CLIENT_TEMPLATE_ATTRIBUTES; renameColumn newColumnName=SCOPE_ID, oldColumnName...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.0.0.CR1','psilva@redhat.com','META-INF/jpa-changelog-authz-4.0.0.CR1.xml','2021-11-01 19:32:39',59,'EXECUTED','7:57960fc0b0f0dd0563ea6f8b2e4a1707','createTable tableName=RESOURCE_SERVER_PERM_TICKET; addPrimaryKey constraintName=CONSTRAINT_FAPMT, tableName=RESOURCE_SERVER_PERM_TICKET; addForeignKeyConstraint baseTableName=RESOURCE_SERVER_PERM_TICKET, constraintName=FK_FRSRHO213XCX4WNKOG82SSPMT...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.0.0.Beta3','psilva@redhat.com','META-INF/jpa-changelog-authz-4.0.0.Beta3.xml','2021-11-01 19:32:39',60,'EXECUTED','7:2b4b8bff39944c7097977cc18dbceb3b','addColumn tableName=RESOURCE_SERVER_POLICY; addColumn tableName=RESOURCE_SERVER_PERM_TICKET; addForeignKeyConstraint baseTableName=RESOURCE_SERVER_PERM_TICKET, constraintName=FK_FRSRPO2128CX4WNKOG82SSRFY, referencedTableName=RESOURCE_SERVER_POLICY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.2.0.Final','mhajas@redhat.com','META-INF/jpa-changelog-authz-4.2.0.Final.xml','2021-11-01 19:32:39',61,'EXECUTED','7:2aa42a964c59cd5b8ca9822340ba33a8','createTable tableName=RESOURCE_URIS; addForeignKeyConstraint baseTableName=RESOURCE_URIS, constraintName=FK_RESOURCE_SERVER_URIS, referencedTableName=RESOURCE_SERVER_RESOURCE; customChange; dropColumn columnName=URI, tableName=RESOURCE_SERVER_RESO...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-4.2.0.Final-KEYCLOAK-9944','hmlnarik@redhat.com','META-INF/jpa-changelog-authz-4.2.0.Final.xml','2021-11-01 19:32:39',62,'EXECUTED','7:9ac9e58545479929ba23f4a3087a0346','addPrimaryKey constraintName=CONSTRAINT_RESOUR_URIS_PK, tableName=RESOURCE_URIS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.2.0-KEYCLOAK-6313','wadahiro@gmail.com','META-INF/jpa-changelog-4.2.0.xml','2021-11-01 19:32:39',63,'EXECUTED','7:14d407c35bc4fe1976867756bcea0c36','addColumn tableName=REQUIRED_ACTION_PROVIDER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.3.0-KEYCLOAK-7984','wadahiro@gmail.com','META-INF/jpa-changelog-4.3.0.xml','2021-11-01 19:32:39',64,'EXECUTED','7:241a8030c748c8548e346adee548fa93','update tableName=REQUIRED_ACTION_PROVIDER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.6.0-KEYCLOAK-7950','psilva@redhat.com','META-INF/jpa-changelog-4.6.0.xml','2021-11-01 19:32:39',65,'EXECUTED','7:7d3182f65a34fcc61e8d23def037dc3f','update tableName=RESOURCE_SERVER_RESOURCE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.6.0-KEYCLOAK-8377','keycloak','META-INF/jpa-changelog-4.6.0.xml','2021-11-01 19:32:39',66,'EXECUTED','7:b30039e00a0b9715d430d1b0636728fa','createTable tableName=ROLE_ATTRIBUTE; addPrimaryKey constraintName=CONSTRAINT_ROLE_ATTRIBUTE_PK, tableName=ROLE_ATTRIBUTE; addForeignKeyConstraint baseTableName=ROLE_ATTRIBUTE, constraintName=FK_ROLE_ATTRIBUTE_ID, referencedTableName=KEYCLOAK_ROLE...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.6.0-KEYCLOAK-8555','gideonray@gmail.com','META-INF/jpa-changelog-4.6.0.xml','2021-11-01 19:32:39',67,'EXECUTED','7:3797315ca61d531780f8e6f82f258159','createIndex indexName=IDX_COMPONENT_PROVIDER_TYPE, tableName=COMPONENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.7.0-KEYCLOAK-1267','sguilhen@redhat.com','META-INF/jpa-changelog-4.7.0.xml','2021-11-01 19:32:39',68,'EXECUTED','7:c7aa4c8d9573500c2d347c1941ff0301','addColumn tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.7.0-KEYCLOAK-7275','keycloak','META-INF/jpa-changelog-4.7.0.xml','2021-11-01 19:32:39',69,'EXECUTED','7:b207faee394fc074a442ecd42185a5dd','renameColumn newColumnName=CREATED_ON, oldColumnName=LAST_SESSION_REFRESH, tableName=OFFLINE_USER_SESSION; addNotNullConstraint columnName=CREATED_ON, tableName=OFFLINE_USER_SESSION; addColumn tableName=OFFLINE_USER_SESSION; customChange; createIn...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('4.8.0-KEYCLOAK-8835','sguilhen@redhat.com','META-INF/jpa-changelog-4.8.0.xml','2021-11-01 19:32:39',70,'EXECUTED','7:ab9a9762faaba4ddfa35514b212c4922','addNotNullConstraint columnName=SSO_MAX_LIFESPAN_REMEMBER_ME, tableName=REALM; addNotNullConstraint columnName=SSO_IDLE_TIMEOUT_REMEMBER_ME, tableName=REALM','',NULL,'3.5.4',NULL,NULL,'5795153174'),('authz-7.0.0-KEYCLOAK-10443','psilva@redhat.com','META-INF/jpa-changelog-authz-7.0.0.xml','2021-11-01 19:32:39',71,'EXECUTED','7:b9710f74515a6ccb51b72dc0d19df8c4','addColumn tableName=RESOURCE_SERVER','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-adding-credential-columns','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',72,'EXECUTED','7:ec9707ae4d4f0b7452fee20128083879','addColumn tableName=CREDENTIAL; addColumn tableName=FED_USER_CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-updating-credential-data-not-oracle-fixed','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',73,'EXECUTED','7:3979a0ae07ac465e920ca696532fc736','update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-updating-credential-data-oracle-fixed','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',74,'MARK_RAN','7:5abfde4c259119d143bd2fbf49ac2bca','update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-credential-cleanup-fixed','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',75,'EXECUTED','7:b48da8c11a3d83ddd6b7d0c8c2219345','dropDefaultValue columnName=COUNTER, tableName=CREDENTIAL; dropDefaultValue columnName=DIGITS, tableName=CREDENTIAL; dropDefaultValue columnName=PERIOD, tableName=CREDENTIAL; dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; dropColumn ...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('8.0.0-resource-tag-support','keycloak','META-INF/jpa-changelog-8.0.0.xml','2021-11-01 19:32:39',76,'EXECUTED','7:a73379915c23bfad3e8f5c6d5c0aa4bd','addColumn tableName=MIGRATION_MODEL; createIndex indexName=IDX_UPDATE_TIME, tableName=MIGRATION_MODEL','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-always-display-client','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',77,'EXECUTED','7:39e0073779aba192646291aa2332493d','addColumn tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-drop-constraints-for-column-increase','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',78,'MARK_RAN','7:81f87368f00450799b4bf42ea0b3ec34','dropUniqueConstraint constraintName=UK_FRSR6T700S9V50BU18WS5PMT, tableName=RESOURCE_SERVER_PERM_TICKET; dropUniqueConstraint constraintName=UK_FRSR6T700S9V50BU18WS5HA6, tableName=RESOURCE_SERVER_RESOURCE; dropPrimaryKey constraintName=CONSTRAINT_O...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-increase-column-size-federated-fk','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',79,'EXECUTED','7:20b37422abb9fb6571c618148f013a15','modifyDataType columnName=CLIENT_ID, tableName=FED_USER_CONSENT; modifyDataType columnName=CLIENT_REALM_CONSTRAINT, tableName=KEYCLOAK_ROLE; modifyDataType columnName=OWNER, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=CLIENT_ID, ta...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.0-recreate-constraints-after-column-increase','keycloak','META-INF/jpa-changelog-9.0.0.xml','2021-11-01 19:32:39',80,'MARK_RAN','7:1970bb6cfb5ee800736b95ad3fb3c78a','addNotNullConstraint columnName=CLIENT_ID, tableName=OFFLINE_CLIENT_SESSION; addNotNullConstraint columnName=OWNER, tableName=RESOURCE_SERVER_PERM_TICKET; addNotNullConstraint columnName=REQUESTER, tableName=RESOURCE_SERVER_PERM_TICKET; addNotNull...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-add-index-to-client.client_id','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',81,'EXECUTED','7:45d9b25fc3b455d522d8dcc10a0f4c80','createIndex indexName=IDX_CLIENT_ID, tableName=CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-KEYCLOAK-12579-drop-constraints','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',82,'MARK_RAN','7:890ae73712bc187a66c2813a724d037f','dropUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-KEYCLOAK-12579-add-not-null-constraint','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',83,'EXECUTED','7:0a211980d27fafe3ff50d19a3a29b538','addNotNullConstraint columnName=PARENT_GROUP, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-KEYCLOAK-12579-recreate-constraints','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',84,'MARK_RAN','7:a161e2ae671a9020fff61e996a207377','addUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP','',NULL,'3.5.4',NULL,NULL,'5795153174'),('9.0.1-add-index-to-events','keycloak','META-INF/jpa-changelog-9.0.1.xml','2021-11-01 19:32:39',85,'EXECUTED','7:01c49302201bdf815b0a18d1f98a55dc','createIndex indexName=IDX_EVENT_TIME, tableName=EVENT_ENTITY','',NULL,'3.5.4',NULL,NULL,'5795153174'),('map-remove-ri','keycloak','META-INF/jpa-changelog-11.0.0.xml','2021-11-01 19:32:39',86,'EXECUTED','7:3dace6b144c11f53f1ad2c0361279b86','dropForeignKeyConstraint baseTableName=REALM, constraintName=FK_TRAF444KK6QRKMS7N56AIWQ5Y; dropForeignKeyConstraint baseTableName=KEYCLOAK_ROLE, constraintName=FK_KJHO5LE2C0RAL09FL8CM9WFW9','',NULL,'3.5.4',NULL,NULL,'5795153174'),('map-remove-ri','keycloak','META-INF/jpa-changelog-12.0.0.xml','2021-11-01 19:32:40',87,'EXECUTED','7:578d0b92077eaf2ab95ad0ec087aa903','dropForeignKeyConstraint baseTableName=REALM_DEFAULT_GROUPS, constraintName=FK_DEF_GROUPS_GROUP; dropForeignKeyConstraint baseTableName=REALM_DEFAULT_ROLES, constraintName=FK_H4WPD7W4HSOOLNI3H0SW7BTJE; dropForeignKeyConstraint baseTableName=CLIENT...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('12.1.0-add-realm-localization-table','keycloak','META-INF/jpa-changelog-12.0.0.xml','2021-11-01 19:32:40',88,'EXECUTED','7:c95abe90d962c57a09ecaee57972835d','createTable tableName=REALM_LOCALIZATIONS; addPrimaryKey tableName=REALM_LOCALIZATIONS','',NULL,'3.5.4',NULL,NULL,'5795153174'),('default-roles','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',89,'EXECUTED','7:f1313bcc2994a5c4dc1062ed6d8282d3','addColumn tableName=REALM; customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('default-roles-cleanup','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',90,'EXECUTED','7:90d763b52eaffebefbcbde55f269508b','dropTable tableName=REALM_DEFAULT_ROLES; dropTable tableName=CLIENT_DEFAULT_ROLES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-KEYCLOAK-16844','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',91,'EXECUTED','7:d554f0cb92b764470dccfa5e0014a7dd','createIndex indexName=IDX_OFFLINE_USS_PRELOAD, tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('map-remove-ri-13.0.0','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',92,'EXECUTED','7:73193e3ab3c35cf0f37ccea3bf783764','dropForeignKeyConstraint baseTableName=DEFAULT_CLIENT_SCOPE, constraintName=FK_R_DEF_CLI_SCOPE_SCOPE; dropForeignKeyConstraint baseTableName=CLIENT_SCOPE_CLIENT, constraintName=FK_C_CLI_SCOPE_SCOPE; dropForeignKeyConstraint baseTableName=CLIENT_SC...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-KEYCLOAK-17992-drop-constraints','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',93,'MARK_RAN','7:90a1e74f92e9cbaa0c5eab80b8a037f3','dropPrimaryKey constraintName=C_CLI_SCOPE_BIND, tableName=CLIENT_SCOPE_CLIENT; dropIndex indexName=IDX_CLSCOPE_CL, tableName=CLIENT_SCOPE_CLIENT; dropIndex indexName=IDX_CL_CLSCOPE, tableName=CLIENT_SCOPE_CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-increase-column-size-federated','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',94,'EXECUTED','7:5b9248f29cd047c200083cc6d8388b16','modifyDataType columnName=CLIENT_ID, tableName=CLIENT_SCOPE_CLIENT; modifyDataType columnName=SCOPE_ID, tableName=CLIENT_SCOPE_CLIENT','',NULL,'3.5.4',NULL,NULL,'5795153174'),('13.0.0-KEYCLOAK-17992-recreate-constraints','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',95,'MARK_RAN','7:64db59e44c374f13955489e8990d17a1','addNotNullConstraint columnName=CLIENT_ID, tableName=CLIENT_SCOPE_CLIENT; addNotNullConstraint columnName=SCOPE_ID, tableName=CLIENT_SCOPE_CLIENT; addPrimaryKey constraintName=C_CLI_SCOPE_BIND, tableName=CLIENT_SCOPE_CLIENT; createIndex indexName=...','',NULL,'3.5.4',NULL,NULL,'5795153174'),('json-string-accomodation-fixed','keycloak','META-INF/jpa-changelog-13.0.0.xml','2021-11-01 19:32:40',96,'EXECUTED','7:329a578cdb43262fff975f0a7f6cda60','addColumn tableName=REALM_ATTRIBUTE; update tableName=REALM_ATTRIBUTE; dropColumn columnName=VALUE, tableName=REALM_ATTRIBUTE; renameColumn newColumnName=VALUE, oldColumnName=VALUE_NEW, tableName=REALM_ATTRIBUTE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-11019','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',97,'EXECUTED','7:fae0de241ac0fd0bbc2b380b85e4f567','createIndex indexName=IDX_OFFLINE_CSS_PRELOAD, tableName=OFFLINE_CLIENT_SESSION; createIndex indexName=IDX_OFFLINE_USS_BY_USER, tableName=OFFLINE_USER_SESSION; createIndex indexName=IDX_OFFLINE_USS_BY_USERSESS, tableName=OFFLINE_USER_SESSION','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',98,'MARK_RAN','7:075d54e9180f49bb0c64ca4218936e81','createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286-revert','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',99,'MARK_RAN','7:06499836520f4f6b3d05e35a59324910','dropIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286-supported-dbs','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',100,'EXECUTED','7:b558ad47ea0e4d3c3514225a49cc0d65','createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('14.0.0-KEYCLOAK-18286-unsupported-dbs','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',101,'MARK_RAN','7:3d2b23076e59c6f70bae703aa01be35b','createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES','',NULL,'3.5.4',NULL,NULL,'5795153174'),('KEYCLOAK-17267-add-index-to-user-attributes','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',102,'EXECUTED','7:1a7f28ff8d9e53aeb879d76ea3d9341a','createIndex indexName=IDX_USER_ATTRIBUTE_NAME, tableName=USER_ATTRIBUTE','',NULL,'3.5.4',NULL,NULL,'5795153174'),('KEYCLOAK-18146-add-saml-art-binding-identifier','keycloak','META-INF/jpa-changelog-14.0.0.xml','2021-11-01 19:32:40',103,'EXECUTED','7:2fd554456fed4a82c698c555c5b751b6','customChange','',NULL,'3.5.4',NULL,NULL,'5795153174'),('15.0.0-KEYCLOAK-18467','keycloak','META-INF/jpa-changelog-15.0.0.xml','2021-11-01 19:32:40',104,'EXECUTED','7:b06356d66c2790ecc2ae54ba0458397a','addColumn tableName=REALM_LOCALIZATIONS; update tableName=REALM_LOCALIZATIONS; dropColumn columnName=TEXTS, tableName=REALM_LOCALIZATIONS; renameColumn newColumnName=TEXTS, oldColumnName=TEXTS_NEW, tableName=REALM_LOCALIZATIONS; addNotNullConstrai...','',NULL,'3.5.4',NULL,NULL,'5795153174');
INSERT INTO databasechangelog VALUES ('1.0.0.Final-KEYCLOAK-5461', 'sthorger@redhat.com', 'META-INF/jpa-changelog-1.0.0.Final.xml', '2025-02-25 13:12:11.661503', 1, 'EXECUTED', '9:6f1016664e21e16d26517a4418f5e3df', 'createTable tableName=APPLICATION_DEFAULT_ROLES; createTable tableName=CLIENT; createTable tableName=CLIENT_SESSION; createTable tableName=CLIENT_SESSION_ROLE; createTable tableName=COMPOSITE_ROLE; createTable tableName=CREDENTIAL; createTable tab...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.0.0.Final-KEYCLOAK-5461', 'sthorger@redhat.com', 'META-INF/db2-jpa-changelog-1.0.0.Final.xml', '2025-02-25 13:12:11.67809', 2, 'MARK_RAN', '9:828775b1596a07d1200ba1d49e5e3941', 'createTable tableName=APPLICATION_DEFAULT_ROLES; createTable tableName=CLIENT; createTable tableName=CLIENT_SESSION; createTable tableName=CLIENT_SESSION_ROLE; createTable tableName=COMPOSITE_ROLE; createTable tableName=CREDENTIAL; createTable tab...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.1.0.Beta1', 'sthorger@redhat.com', 'META-INF/jpa-changelog-1.1.0.Beta1.xml', '2025-02-25 13:12:11.720836', 3, 'EXECUTED', '9:5f090e44a7d595883c1fb61f4b41fd38', 'delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=CLIENT_ATTRIBUTES; createTable tableName=CLIENT_SESSION_NOTE; createTable tableName=APP_NODE_REGISTRATIONS; addColumn table...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.1.0.Final', 'sthorger@redhat.com', 'META-INF/jpa-changelog-1.1.0.Final.xml', '2025-02-25 13:12:11.724389', 4, 'EXECUTED', '9:c07e577387a3d2c04d1adc9aaad8730e', 'renameColumn newColumnName=EVENT_TIME, oldColumnName=TIME, tableName=EVENT_ENTITY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.2.0.Beta1', 'psilva@redhat.com', 'META-INF/jpa-changelog-1.2.0.Beta1.xml', '2025-02-25 13:12:11.831874', 5, 'EXECUTED', '9:b68ce996c655922dbcd2fe6b6ae72686', 'delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=PROTOCOL_MAPPER; createTable tableName=PROTOCOL_MAPPER_CONFIG; createTable tableName=...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.2.0.Beta1', 'psilva@redhat.com', 'META-INF/db2-jpa-changelog-1.2.0.Beta1.xml', '2025-02-25 13:12:11.83815', 6, 'MARK_RAN', '9:543b5c9989f024fe35c6f6c5a97de88e', 'delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION; createTable tableName=PROTOCOL_MAPPER; createTable tableName=PROTOCOL_MAPPER_CONFIG; createTable tableName=...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.2.0.RC1', 'bburke@redhat.com', 'META-INF/jpa-changelog-1.2.0.CR1.xml', '2025-02-25 13:12:11.917197', 7, 'EXECUTED', '9:765afebbe21cf5bbca048e632df38336', 'delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=MIGRATION_MODEL; createTable tableName=IDENTITY_P...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.2.0.RC1', 'bburke@redhat.com', 'META-INF/db2-jpa-changelog-1.2.0.CR1.xml', '2025-02-25 13:12:11.924499', 8, 'MARK_RAN', '9:db4a145ba11a6fdaefb397f6dbf829a1', 'delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=MIGRATION_MODEL; createTable tableName=IDENTITY_P...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.2.0.Final', 'keycloak', 'META-INF/jpa-changelog-1.2.0.Final.xml', '2025-02-25 13:12:11.930916', 9, 'EXECUTED', '9:9d05c7be10cdb873f8bcb41bc3a8ab23', 'update tableName=CLIENT; update tableName=CLIENT; update tableName=CLIENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.3.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-1.3.0.xml', '2025-02-25 13:12:12.039116', 10, 'EXECUTED', '9:18593702353128d53111f9b1ff0b82b8', 'delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete tableName=USER_SESSION; createTable tableName=ADMI...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.4.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-1.4.0.xml', '2025-02-25 13:12:12.089129', 11, 'EXECUTED', '9:6122efe5f090e41a85c0f1c9e52cbb62', 'delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.4.0', 'bburke@redhat.com', 'META-INF/db2-jpa-changelog-1.4.0.xml', '2025-02-25 13:12:12.092642', 12, 'MARK_RAN', '9:e1ff28bf7568451453f844c5d54bb0b5', 'delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.5.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-1.5.0.xml', '2025-02-25 13:12:12.114437', 13, 'EXECUTED', '9:7af32cd8957fbc069f796b61217483fd', 'delete tableName=CLIENT_SESSION_AUTH_STATUS; delete tableName=CLIENT_SESSION_ROLE; delete tableName=CLIENT_SESSION_PROT_MAPPER; delete tableName=CLIENT_SESSION_NOTE; delete tableName=CLIENT_SESSION; delete tableName=USER_SESSION_NOTE; delete table...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.6.1_from15', 'mposolda@redhat.com', 'META-INF/jpa-changelog-1.6.1.xml', '2025-02-25 13:12:12.129013', 14, 'EXECUTED', '9:6005e15e84714cd83226bf7879f54190', 'addColumn tableName=REALM; addColumn tableName=KEYCLOAK_ROLE; addColumn tableName=CLIENT; createTable tableName=OFFLINE_USER_SESSION; createTable tableName=OFFLINE_CLIENT_SESSION; addPrimaryKey constraintName=CONSTRAINT_OFFL_US_SES_PK2, tableName=...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.6.1_from16-pre', 'mposolda@redhat.com', 'META-INF/jpa-changelog-1.6.1.xml', '2025-02-25 13:12:12.131835', 15, 'MARK_RAN', '9:bf656f5a2b055d07f314431cae76f06c', 'delete tableName=OFFLINE_CLIENT_SESSION; delete tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.6.1_from16', 'mposolda@redhat.com', 'META-INF/jpa-changelog-1.6.1.xml', '2025-02-25 13:12:12.134179', 16, 'MARK_RAN', '9:f8dadc9284440469dcf71e25ca6ab99b', 'dropPrimaryKey constraintName=CONSTRAINT_OFFLINE_US_SES_PK, tableName=OFFLINE_USER_SESSION; dropPrimaryKey constraintName=CONSTRAINT_OFFLINE_CL_SES_PK, tableName=OFFLINE_CLIENT_SESSION; addColumn tableName=OFFLINE_USER_SESSION; update tableName=OF...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.6.1', 'mposolda@redhat.com', 'META-INF/jpa-changelog-1.6.1.xml', '2025-02-25 13:12:12.137324', 17, 'EXECUTED', '9:d41d8cd98f00b204e9800998ecf8427e', 'empty', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.7.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-1.7.0.xml', '2025-02-25 13:12:12.17413', 18, 'EXECUTED', '9:3368ff0be4c2855ee2dd9ca813b38d8e', 'createTable tableName=KEYCLOAK_GROUP; createTable tableName=GROUP_ROLE_MAPPING; createTable tableName=GROUP_ATTRIBUTE; createTable tableName=USER_GROUP_MEMBERSHIP; createTable tableName=REALM_DEFAULT_GROUPS; addColumn tableName=IDENTITY_PROVIDER; ...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.8.0', 'mposolda@redhat.com', 'META-INF/jpa-changelog-1.8.0.xml', '2025-02-25 13:12:12.227478', 19, 'EXECUTED', '9:8ac2fb5dd030b24c0570a763ed75ed20', 'addColumn tableName=IDENTITY_PROVIDER; createTable tableName=CLIENT_TEMPLATE; createTable tableName=CLIENT_TEMPLATE_ATTRIBUTES; createTable tableName=TEMPLATE_SCOPE_MAPPING; dropNotNullConstraint columnName=CLIENT_ID, tableName=PROTOCOL_MAPPER; ad...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.8.0-2', 'keycloak', 'META-INF/jpa-changelog-1.8.0.xml', '2025-02-25 13:12:12.232206', 20, 'EXECUTED', '9:f91ddca9b19743db60e3057679810e6c', 'dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; update tableName=CREDENTIAL', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.8.0', 'mposolda@redhat.com', 'META-INF/db2-jpa-changelog-1.8.0.xml', '2025-02-25 13:12:12.235679', 21, 'MARK_RAN', '9:831e82914316dc8a57dc09d755f23c51', 'addColumn tableName=IDENTITY_PROVIDER; createTable tableName=CLIENT_TEMPLATE; createTable tableName=CLIENT_TEMPLATE_ATTRIBUTES; createTable tableName=TEMPLATE_SCOPE_MAPPING; dropNotNullConstraint columnName=CLIENT_ID, tableName=PROTOCOL_MAPPER; ad...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.8.0-2', 'keycloak', 'META-INF/db2-jpa-changelog-1.8.0.xml', '2025-02-25 13:12:12.238566', 22, 'MARK_RAN', '9:f91ddca9b19743db60e3057679810e6c', 'dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; update tableName=CREDENTIAL', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.9.0', 'mposolda@redhat.com', 'META-INF/jpa-changelog-1.9.0.xml', '2025-02-25 13:12:12.268407', 23, 'EXECUTED', '9:bc3d0f9e823a69dc21e23e94c7a94bb1', 'update tableName=REALM; update tableName=REALM; update tableName=REALM; update tableName=REALM; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=REALM; update tableName=REALM; customChange; dr...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.9.1', 'keycloak', 'META-INF/jpa-changelog-1.9.1.xml', '2025-02-25 13:12:12.274568', 24, 'EXECUTED', '9:c9999da42f543575ab790e76439a2679', 'modifyDataType columnName=PRIVATE_KEY, tableName=REALM; modifyDataType columnName=PUBLIC_KEY, tableName=REALM; modifyDataType columnName=CERTIFICATE, tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.9.1', 'keycloak', 'META-INF/db2-jpa-changelog-1.9.1.xml', '2025-02-25 13:12:12.275985', 25, 'MARK_RAN', '9:0d6c65c6f58732d81569e77b10ba301d', 'modifyDataType columnName=PRIVATE_KEY, tableName=REALM; modifyDataType columnName=CERTIFICATE, tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('1.9.2', 'keycloak', 'META-INF/jpa-changelog-1.9.2.xml', '2025-02-25 13:12:12.28928', 26, 'EXECUTED', '9:fc576660fc016ae53d2d4778d84d86d0', 'createIndex indexName=IDX_USER_EMAIL, tableName=USER_ENTITY; createIndex indexName=IDX_USER_ROLE_MAPPING, tableName=USER_ROLE_MAPPING; createIndex indexName=IDX_USER_GROUP_MAPPING, tableName=USER_GROUP_MEMBERSHIP; createIndex indexName=IDX_USER_CO...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-2.0.0', 'psilva@redhat.com', 'META-INF/jpa-changelog-authz-2.0.0.xml', '2025-02-25 13:12:12.338249', 27, 'EXECUTED', '9:43ed6b0da89ff77206289e87eaa9c024', 'createTable tableName=RESOURCE_SERVER; addPrimaryKey constraintName=CONSTRAINT_FARS, tableName=RESOURCE_SERVER; addUniqueConstraint constraintName=UK_AU8TT6T700S9V50BU18WS5HA6, tableName=RESOURCE_SERVER; createTable tableName=RESOURCE_SERVER_RESOU...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-2.5.1', 'psilva@redhat.com', 'META-INF/jpa-changelog-authz-2.5.1.xml', '2025-02-25 13:12:12.341358', 28, 'EXECUTED', '9:44bae577f551b3738740281eceb4ea70', 'update tableName=RESOURCE_SERVER_POLICY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.1.0-KEYCLOAK-5461', 'bburke@redhat.com', 'META-INF/jpa-changelog-2.1.0.xml', '2025-02-25 13:12:12.380808', 29, 'EXECUTED', '9:bd88e1f833df0420b01e114533aee5e8', 'createTable tableName=BROKER_LINK; createTable tableName=FED_USER_ATTRIBUTE; createTable tableName=FED_USER_CONSENT; createTable tableName=FED_USER_CONSENT_ROLE; createTable tableName=FED_USER_CONSENT_PROT_MAPPER; createTable tableName=FED_USER_CR...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.2.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-2.2.0.xml', '2025-02-25 13:12:12.39388', 30, 'EXECUTED', '9:a7022af5267f019d020edfe316ef4371', 'addColumn tableName=ADMIN_EVENT_ENTITY; createTable tableName=CREDENTIAL_ATTRIBUTE; createTable tableName=FED_CREDENTIAL_ATTRIBUTE; modifyDataType columnName=VALUE, tableName=CREDENTIAL; addForeignKeyConstraint baseTableName=FED_CREDENTIAL_ATTRIBU...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.3.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-2.3.0.xml', '2025-02-25 13:12:12.4396', 31, 'EXECUTED', '9:fc155c394040654d6a79227e56f5e25a', 'createTable tableName=FEDERATED_USER; addPrimaryKey constraintName=CONSTR_FEDERATED_USER, tableName=FEDERATED_USER; dropDefaultValue columnName=TOTP, tableName=USER_ENTITY; dropColumn columnName=TOTP, tableName=USER_ENTITY; addColumn tableName=IDE...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.4.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-2.4.0.xml', '2025-02-25 13:12:12.448497', 32, 'EXECUTED', '9:eac4ffb2a14795e5dc7b426063e54d88', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.5.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-2.5.0.xml', '2025-02-25 13:12:12.457994', 33, 'EXECUTED', '9:54937c05672568c4c64fc9524c1e9462', 'customChange; modifyDataType columnName=USER_ID, tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.5.0-unicode-oracle', 'hmlnarik@redhat.com', 'META-INF/jpa-changelog-2.5.0.xml', '2025-02-25 13:12:12.459896', 34, 'MARK_RAN', '9:3a32bace77c84d7678d035a7f5a8084e', 'modifyDataType columnName=DESCRIPTION, tableName=AUTHENTICATION_FLOW; modifyDataType columnName=DESCRIPTION, tableName=CLIENT_TEMPLATE; modifyDataType columnName=DESCRIPTION, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=DESCRIPTION,...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.5.0-unicode-other-dbs', 'hmlnarik@redhat.com', 'META-INF/jpa-changelog-2.5.0.xml', '2025-02-25 13:12:12.485883', 35, 'EXECUTED', '9:33d72168746f81f98ae3a1e8e0ca3554', 'modifyDataType columnName=DESCRIPTION, tableName=AUTHENTICATION_FLOW; modifyDataType columnName=DESCRIPTION, tableName=CLIENT_TEMPLATE; modifyDataType columnName=DESCRIPTION, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=DESCRIPTION,...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.5.0-duplicate-email-support', 'slawomir@dabek.name', 'META-INF/jpa-changelog-2.5.0.xml', '2025-02-25 13:12:12.491902', 36, 'EXECUTED', '9:61b6d3d7a4c0e0024b0c839da283da0c', 'addColumn tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.5.0-unique-group-names', 'hmlnarik@redhat.com', 'META-INF/jpa-changelog-2.5.0.xml', '2025-02-25 13:12:12.495487', 37, 'EXECUTED', '9:8dcac7bdf7378e7d823cdfddebf72fda', 'addUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('2.5.1', 'bburke@redhat.com', 'META-INF/jpa-changelog-2.5.1.xml', '2025-02-25 13:12:12.498458', 38, 'EXECUTED', '9:a2b870802540cb3faa72098db5388af3', 'addColumn tableName=FED_USER_CONSENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.0.0', 'bburke@redhat.com', 'META-INF/jpa-changelog-3.0.0.xml', '2025-02-25 13:12:12.501202', 39, 'EXECUTED', '9:132a67499ba24bcc54fb5cbdcfe7e4c0', 'addColumn tableName=IDENTITY_PROVIDER', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.2.0-fix', 'keycloak', 'META-INF/jpa-changelog-3.2.0.xml', '2025-02-25 13:12:12.502127', 40, 'MARK_RAN', '9:938f894c032f5430f2b0fafb1a243462', 'addNotNullConstraint columnName=REALM_ID, tableName=CLIENT_INITIAL_ACCESS', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.2.0-fix-with-keycloak-5416', 'keycloak', 'META-INF/jpa-changelog-3.2.0.xml', '2025-02-25 13:12:12.504564', 41, 'MARK_RAN', '9:845c332ff1874dc5d35974b0babf3006', 'dropIndex indexName=IDX_CLIENT_INIT_ACC_REALM, tableName=CLIENT_INITIAL_ACCESS; addNotNullConstraint columnName=REALM_ID, tableName=CLIENT_INITIAL_ACCESS; createIndex indexName=IDX_CLIENT_INIT_ACC_REALM, tableName=CLIENT_INITIAL_ACCESS', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.2.0-fix-offline-sessions', 'hmlnarik', 'META-INF/jpa-changelog-3.2.0.xml', '2025-02-25 13:12:12.511225', 42, 'EXECUTED', '9:fc86359c079781adc577c5a217e4d04c', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.2.0-fixed', 'keycloak', 'META-INF/jpa-changelog-3.2.0.xml', '2025-02-25 13:12:12.566812', 43, 'EXECUTED', '9:59a64800e3c0d09b825f8a3b444fa8f4', 'addColumn tableName=REALM; dropPrimaryKey constraintName=CONSTRAINT_OFFL_CL_SES_PK2, tableName=OFFLINE_CLIENT_SESSION; dropColumn columnName=CLIENT_SESSION_ID, tableName=OFFLINE_CLIENT_SESSION; addPrimaryKey constraintName=CONSTRAINT_OFFL_CL_SES_P...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.3.0', 'keycloak', 'META-INF/jpa-changelog-3.3.0.xml', '2025-02-25 13:12:12.570283', 44, 'EXECUTED', '9:d48d6da5c6ccf667807f633fe489ce88', 'addColumn tableName=USER_ENTITY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-3.4.0.CR1-resource-server-pk-change-part1', 'glavoie@gmail.com', 'META-INF/jpa-changelog-authz-3.4.0.CR1.xml', '2025-02-25 13:12:12.573738', 45, 'EXECUTED', '9:dde36f7973e80d71fceee683bc5d2951', 'addColumn tableName=RESOURCE_SERVER_POLICY; addColumn tableName=RESOURCE_SERVER_RESOURCE; addColumn tableName=RESOURCE_SERVER_SCOPE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-3.4.0.CR1-resource-server-pk-change-part2-KEYCLOAK-6095', 'hmlnarik@redhat.com', 'META-INF/jpa-changelog-authz-3.4.0.CR1.xml', '2025-02-25 13:12:12.579907', 46, 'EXECUTED', '9:b855e9b0a406b34fa323235a0cf4f640', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-3.4.0.CR1-resource-server-pk-change-part3-fixed', 'glavoie@gmail.com', 'META-INF/jpa-changelog-authz-3.4.0.CR1.xml', '2025-02-25 13:12:12.581399', 47, 'MARK_RAN', '9:51abbacd7b416c50c4421a8cabf7927e', 'dropIndex indexName=IDX_RES_SERV_POL_RES_SERV, tableName=RESOURCE_SERVER_POLICY; dropIndex indexName=IDX_RES_SRV_RES_RES_SRV, tableName=RESOURCE_SERVER_RESOURCE; dropIndex indexName=IDX_RES_SRV_SCOPE_RES_SRV, tableName=RESOURCE_SERVER_SCOPE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-3.4.0.CR1-resource-server-pk-change-part3-fixed-nodropindex', 'glavoie@gmail.com', 'META-INF/jpa-changelog-authz-3.4.0.CR1.xml', '2025-02-25 13:12:12.619409', 48, 'EXECUTED', '9:bdc99e567b3398bac83263d375aad143', 'addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, tableName=RESOURCE_SERVER_POLICY; addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, tableName=RESOURCE_SERVER_RESOURCE; addNotNullConstraint columnName=RESOURCE_SERVER_CLIENT_ID, ...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authn-3.4.0.CR1-refresh-token-max-reuse', 'glavoie@gmail.com', 'META-INF/jpa-changelog-authz-3.4.0.CR1.xml', '2025-02-25 13:12:12.622931', 49, 'EXECUTED', '9:d198654156881c46bfba39abd7769e69', 'addColumn tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.4.0', 'keycloak', 'META-INF/jpa-changelog-3.4.0.xml', '2025-02-25 13:12:12.669535', 50, 'EXECUTED', '9:cfdd8736332ccdd72c5256ccb42335db', 'addPrimaryKey constraintName=CONSTRAINT_REALM_DEFAULT_ROLES, tableName=REALM_DEFAULT_ROLES; addPrimaryKey constraintName=CONSTRAINT_COMPOSITE_ROLE, tableName=COMPOSITE_ROLE; addPrimaryKey constraintName=CONSTR_REALM_DEFAULT_GROUPS, tableName=REALM...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.4.0-KEYCLOAK-5230', 'hmlnarik@redhat.com', 'META-INF/jpa-changelog-3.4.0.xml', '2025-02-25 13:12:12.690466', 51, 'EXECUTED', '9:7c84de3d9bd84d7f077607c1a4dcb714', 'createIndex indexName=IDX_FU_ATTRIBUTE, tableName=FED_USER_ATTRIBUTE; createIndex indexName=IDX_FU_CONSENT, tableName=FED_USER_CONSENT; createIndex indexName=IDX_FU_CONSENT_RU, tableName=FED_USER_CONSENT; createIndex indexName=IDX_FU_CREDENTIAL, t...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.4.1', 'psilva@redhat.com', 'META-INF/jpa-changelog-3.4.1.xml', '2025-02-25 13:12:12.699901', 52, 'EXECUTED', '9:5a6bb36cbefb6a9d6928452c0852af2d', 'modifyDataType columnName=VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.4.2', 'keycloak', 'META-INF/jpa-changelog-3.4.2.xml', '2025-02-25 13:12:12.702024', 53, 'EXECUTED', '9:8f23e334dbc59f82e0a328373ca6ced0', 'update tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('3.4.2-KEYCLOAK-5172', 'mkanis@redhat.com', 'META-INF/jpa-changelog-3.4.2.xml', '2025-02-25 13:12:12.707119', 54, 'EXECUTED', '9:9156214268f09d970cdf0e1564d866af', 'update tableName=CLIENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.0.0-KEYCLOAK-6335', 'bburke@redhat.com', 'META-INF/jpa-changelog-4.0.0.xml', '2025-02-25 13:12:12.72063', 55, 'EXECUTED', '9:db806613b1ed154826c02610b7dbdf74', 'createTable tableName=CLIENT_AUTH_FLOW_BINDINGS; addPrimaryKey constraintName=C_CLI_FLOW_BIND, tableName=CLIENT_AUTH_FLOW_BINDINGS', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.0.0-CLEANUP-UNUSED-TABLE', 'bburke@redhat.com', 'META-INF/jpa-changelog-4.0.0.xml', '2025-02-25 13:12:12.733995', 56, 'EXECUTED', '9:229a041fb72d5beac76bb94a5fa709de', 'dropTable tableName=CLIENT_IDENTITY_PROV_MAPPING', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.0.0-KEYCLOAK-6228', 'bburke@redhat.com', 'META-INF/jpa-changelog-4.0.0.xml', '2025-02-25 13:12:12.764587', 57, 'EXECUTED', '9:079899dade9c1e683f26b2aa9ca6ff04', 'dropUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHOGM8UEWRT, tableName=USER_CONSENT; dropNotNullConstraint columnName=CLIENT_ID, tableName=USER_CONSENT; addColumn tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHO...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.0.0-KEYCLOAK-5579-fixed', 'mposolda@redhat.com', 'META-INF/jpa-changelog-4.0.0.xml', '2025-02-25 13:12:12.862426', 58, 'EXECUTED', '9:139b79bcbbfe903bb1c2d2a4dbf001d9', 'dropForeignKeyConstraint baseTableName=CLIENT_TEMPLATE_ATTRIBUTES, constraintName=FK_CL_TEMPL_ATTR_TEMPL; renameTable newTableName=CLIENT_SCOPE_ATTRIBUTES, oldTableName=CLIENT_TEMPLATE_ATTRIBUTES; renameColumn newColumnName=SCOPE_ID, oldColumnName...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-4.0.0.CR1', 'psilva@redhat.com', 'META-INF/jpa-changelog-authz-4.0.0.CR1.xml', '2025-02-25 13:12:12.881506', 59, 'EXECUTED', '9:b55738ad889860c625ba2bf483495a04', 'createTable tableName=RESOURCE_SERVER_PERM_TICKET; addPrimaryKey constraintName=CONSTRAINT_FAPMT, tableName=RESOURCE_SERVER_PERM_TICKET; addForeignKeyConstraint baseTableName=RESOURCE_SERVER_PERM_TICKET, constraintName=FK_FRSRHO213XCX4WNKOG82SSPMT...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-4.0.0.Beta3', 'psilva@redhat.com', 'META-INF/jpa-changelog-authz-4.0.0.Beta3.xml', '2025-02-25 13:12:12.887203', 60, 'EXECUTED', '9:e0057eac39aa8fc8e09ac6cfa4ae15fe', 'addColumn tableName=RESOURCE_SERVER_POLICY; addColumn tableName=RESOURCE_SERVER_PERM_TICKET; addForeignKeyConstraint baseTableName=RESOURCE_SERVER_PERM_TICKET, constraintName=FK_FRSRPO2128CX4WNKOG82SSRFY, referencedTableName=RESOURCE_SERVER_POLICY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-4.2.0.Final', 'mhajas@redhat.com', 'META-INF/jpa-changelog-authz-4.2.0.Final.xml', '2025-02-25 13:12:12.897139', 61, 'EXECUTED', '9:42a33806f3a0443fe0e7feeec821326c', 'createTable tableName=RESOURCE_URIS; addForeignKeyConstraint baseTableName=RESOURCE_URIS, constraintName=FK_RESOURCE_SERVER_URIS, referencedTableName=RESOURCE_SERVER_RESOURCE; customChange; dropColumn columnName=URI, tableName=RESOURCE_SERVER_RESO...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-4.2.0.Final-KEYCLOAK-9944', 'hmlnarik@redhat.com', 'META-INF/jpa-changelog-authz-4.2.0.Final.xml', '2025-02-25 13:12:12.900203', 62, 'EXECUTED', '9:9968206fca46eecc1f51db9c024bfe56', 'addPrimaryKey constraintName=CONSTRAINT_RESOUR_URIS_PK, tableName=RESOURCE_URIS', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.2.0-KEYCLOAK-6313', 'wadahiro@gmail.com', 'META-INF/jpa-changelog-4.2.0.xml', '2025-02-25 13:12:12.902995', 63, 'EXECUTED', '9:92143a6daea0a3f3b8f598c97ce55c3d', 'addColumn tableName=REQUIRED_ACTION_PROVIDER', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.3.0-KEYCLOAK-7984', 'wadahiro@gmail.com', 'META-INF/jpa-changelog-4.3.0.xml', '2025-02-25 13:12:12.905376', 64, 'EXECUTED', '9:82bab26a27195d889fb0429003b18f40', 'update tableName=REQUIRED_ACTION_PROVIDER', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.6.0-KEYCLOAK-7950', 'psilva@redhat.com', 'META-INF/jpa-changelog-4.6.0.xml', '2025-02-25 13:12:12.907442', 65, 'EXECUTED', '9:e590c88ddc0b38b0ae4249bbfcb5abc3', 'update tableName=RESOURCE_SERVER_RESOURCE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.6.0-KEYCLOAK-8377', 'keycloak', 'META-INF/jpa-changelog-4.6.0.xml', '2025-02-25 13:12:12.914879', 66, 'EXECUTED', '9:5c1f475536118dbdc38d5d7977950cc0', 'createTable tableName=ROLE_ATTRIBUTE; addPrimaryKey constraintName=CONSTRAINT_ROLE_ATTRIBUTE_PK, tableName=ROLE_ATTRIBUTE; addForeignKeyConstraint baseTableName=ROLE_ATTRIBUTE, constraintName=FK_ROLE_ATTRIBUTE_ID, referencedTableName=KEYCLOAK_ROLE...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.6.0-KEYCLOAK-8555', 'gideonray@gmail.com', 'META-INF/jpa-changelog-4.6.0.xml', '2025-02-25 13:12:12.918213', 67, 'EXECUTED', '9:e7c9f5f9c4d67ccbbcc215440c718a17', 'createIndex indexName=IDX_COMPONENT_PROVIDER_TYPE, tableName=COMPONENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.7.0-KEYCLOAK-1267', 'sguilhen@redhat.com', 'META-INF/jpa-changelog-4.7.0.xml', '2025-02-25 13:12:12.921866', 68, 'EXECUTED', '9:88e0bfdda924690d6f4e430c53447dd5', 'addColumn tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.7.0-KEYCLOAK-7275', 'keycloak', 'META-INF/jpa-changelog-4.7.0.xml', '2025-02-25 13:12:12.930773', 69, 'EXECUTED', '9:f53177f137e1c46b6a88c59ec1cb5218', 'renameColumn newColumnName=CREATED_ON, oldColumnName=LAST_SESSION_REFRESH, tableName=OFFLINE_USER_SESSION; addNotNullConstraint columnName=CREATED_ON, tableName=OFFLINE_USER_SESSION; addColumn tableName=OFFLINE_USER_SESSION; customChange; createIn...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('4.8.0-KEYCLOAK-8835', 'sguilhen@redhat.com', 'META-INF/jpa-changelog-4.8.0.xml', '2025-02-25 13:12:12.934124', 70, 'EXECUTED', '9:a74d33da4dc42a37ec27121580d1459f', 'addNotNullConstraint columnName=SSO_MAX_LIFESPAN_REMEMBER_ME, tableName=REALM; addNotNullConstraint columnName=SSO_IDLE_TIMEOUT_REMEMBER_ME, tableName=REALM', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('authz-7.0.0-KEYCLOAK-10443', 'psilva@redhat.com', 'META-INF/jpa-changelog-authz-7.0.0.xml', '2025-02-25 13:12:12.936608', 71, 'EXECUTED', '9:fd4ade7b90c3b67fae0bfcfcb42dfb5f', 'addColumn tableName=RESOURCE_SERVER', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('8.0.0-adding-credential-columns', 'keycloak', 'META-INF/jpa-changelog-8.0.0.xml', '2025-02-25 13:12:12.941478', 72, 'EXECUTED', '9:aa072ad090bbba210d8f18781b8cebf4', 'addColumn tableName=CREDENTIAL; addColumn tableName=FED_USER_CREDENTIAL', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('8.0.0-updating-credential-data-not-oracle-fixed', 'keycloak', 'META-INF/jpa-changelog-8.0.0.xml', '2025-02-25 13:12:12.946304', 73, 'EXECUTED', '9:1ae6be29bab7c2aa376f6983b932be37', 'update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('8.0.0-updating-credential-data-oracle-fixed', 'keycloak', 'META-INF/jpa-changelog-8.0.0.xml', '2025-02-25 13:12:12.948808', 74, 'MARK_RAN', '9:14706f286953fc9a25286dbd8fb30d97', 'update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL; update tableName=FED_USER_CREDENTIAL', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('8.0.0-credential-cleanup-fixed', 'keycloak', 'META-INF/jpa-changelog-8.0.0.xml', '2025-02-25 13:12:12.980896', 75, 'EXECUTED', '9:2b9cc12779be32c5b40e2e67711a218b', 'dropDefaultValue columnName=COUNTER, tableName=CREDENTIAL; dropDefaultValue columnName=DIGITS, tableName=CREDENTIAL; dropDefaultValue columnName=PERIOD, tableName=CREDENTIAL; dropDefaultValue columnName=ALGORITHM, tableName=CREDENTIAL; dropColumn ...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('8.0.0-resource-tag-support', 'keycloak', 'META-INF/jpa-changelog-8.0.0.xml', '2025-02-25 13:12:12.987081', 76, 'EXECUTED', '9:91fa186ce7a5af127a2d7a91ee083cc5', 'addColumn tableName=MIGRATION_MODEL; createIndex indexName=IDX_UPDATE_TIME, tableName=MIGRATION_MODEL', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.0-always-display-client', 'keycloak', 'META-INF/jpa-changelog-9.0.0.xml', '2025-02-25 13:12:12.989947', 77, 'EXECUTED', '9:6335e5c94e83a2639ccd68dd24e2e5ad', 'addColumn tableName=CLIENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.0-drop-constraints-for-column-increase', 'keycloak', 'META-INF/jpa-changelog-9.0.0.xml', '2025-02-25 13:12:12.990915', 78, 'MARK_RAN', '9:6bdb5658951e028bfe16fa0a8228b530', 'dropUniqueConstraint constraintName=UK_FRSR6T700S9V50BU18WS5PMT, tableName=RESOURCE_SERVER_PERM_TICKET; dropUniqueConstraint constraintName=UK_FRSR6T700S9V50BU18WS5HA6, tableName=RESOURCE_SERVER_RESOURCE; dropPrimaryKey constraintName=CONSTRAINT_O...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.0-increase-column-size-federated-fk', 'keycloak', 'META-INF/jpa-changelog-9.0.0.xml', '2025-02-25 13:12:13.008459', 79, 'EXECUTED', '9:d5bc15a64117ccad481ce8792d4c608f', 'modifyDataType columnName=CLIENT_ID, tableName=FED_USER_CONSENT; modifyDataType columnName=CLIENT_REALM_CONSTRAINT, tableName=KEYCLOAK_ROLE; modifyDataType columnName=OWNER, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=CLIENT_ID, ta...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.0-recreate-constraints-after-column-increase', 'keycloak', 'META-INF/jpa-changelog-9.0.0.xml', '2025-02-25 13:12:13.009783', 80, 'MARK_RAN', '9:077cba51999515f4d3e7ad5619ab592c', 'addNotNullConstraint columnName=CLIENT_ID, tableName=OFFLINE_CLIENT_SESSION; addNotNullConstraint columnName=OWNER, tableName=RESOURCE_SERVER_PERM_TICKET; addNotNullConstraint columnName=REQUESTER, tableName=RESOURCE_SERVER_PERM_TICKET; addNotNull...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.1-add-index-to-client.client_id', 'keycloak', 'META-INF/jpa-changelog-9.0.1.xml', '2025-02-25 13:12:13.013843', 81, 'EXECUTED', '9:be969f08a163bf47c6b9e9ead8ac2afb', 'createIndex indexName=IDX_CLIENT_ID, tableName=CLIENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.1-KEYCLOAK-12579-drop-constraints', 'keycloak', 'META-INF/jpa-changelog-9.0.1.xml', '2025-02-25 13:12:13.015477', 82, 'MARK_RAN', '9:6d3bb4408ba5a72f39bd8a0b301ec6e3', 'dropUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.1-KEYCLOAK-12579-add-not-null-constraint', 'keycloak', 'META-INF/jpa-changelog-9.0.1.xml', '2025-02-25 13:12:13.02073', 83, 'EXECUTED', '9:966bda61e46bebf3cc39518fbed52fa7', 'addNotNullConstraint columnName=PARENT_GROUP, tableName=KEYCLOAK_GROUP', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.1-KEYCLOAK-12579-recreate-constraints', 'keycloak', 'META-INF/jpa-changelog-9.0.1.xml', '2025-02-25 13:12:13.022589', 84, 'MARK_RAN', '9:8dcac7bdf7378e7d823cdfddebf72fda', 'addUniqueConstraint constraintName=SIBLING_NAMES, tableName=KEYCLOAK_GROUP', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('9.0.1-add-index-to-events', 'keycloak', 'META-INF/jpa-changelog-9.0.1.xml', '2025-02-25 13:12:13.026294', 85, 'EXECUTED', '9:7d93d602352a30c0c317e6a609b56599', 'createIndex indexName=IDX_EVENT_TIME, tableName=EVENT_ENTITY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('map-remove-ri', 'keycloak', 'META-INF/jpa-changelog-11.0.0.xml', '2025-02-25 13:12:13.031392', 86, 'EXECUTED', '9:71c5969e6cdd8d7b6f47cebc86d37627', 'dropForeignKeyConstraint baseTableName=REALM, constraintName=FK_TRAF444KK6QRKMS7N56AIWQ5Y; dropForeignKeyConstraint baseTableName=KEYCLOAK_ROLE, constraintName=FK_KJHO5LE2C0RAL09FL8CM9WFW9', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('map-remove-ri', 'keycloak', 'META-INF/jpa-changelog-12.0.0.xml', '2025-02-25 13:12:13.042618', 87, 'EXECUTED', '9:a9ba7d47f065f041b7da856a81762021', 'dropForeignKeyConstraint baseTableName=REALM_DEFAULT_GROUPS, constraintName=FK_DEF_GROUPS_GROUP; dropForeignKeyConstraint baseTableName=REALM_DEFAULT_ROLES, constraintName=FK_H4WPD7W4HSOOLNI3H0SW7BTJE; dropForeignKeyConstraint baseTableName=CLIENT...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('12.1.0-add-realm-localization-table', 'keycloak', 'META-INF/jpa-changelog-12.0.0.xml', '2025-02-25 13:12:13.047817', 88, 'EXECUTED', '9:fffabce2bc01e1a8f5110d5278500065', 'createTable tableName=REALM_LOCALIZATIONS; addPrimaryKey tableName=REALM_LOCALIZATIONS', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('default-roles', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.055705', 89, 'EXECUTED', '9:fa8a5b5445e3857f4b010bafb5009957', 'addColumn tableName=REALM; customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('default-roles-cleanup', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.073622', 90, 'EXECUTED', '9:67ac3241df9a8582d591c5ed87125f39', 'dropTable tableName=REALM_DEFAULT_ROLES; dropTable tableName=CLIENT_DEFAULT_ROLES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('13.0.0-KEYCLOAK-16844', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.07886', 91, 'EXECUTED', '9:ad1194d66c937e3ffc82386c050ba089', 'createIndex indexName=IDX_OFFLINE_USS_PRELOAD, tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('map-remove-ri-13.0.0', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.089845', 92, 'EXECUTED', '9:d9be619d94af5a2f5d07b9f003543b91', 'dropForeignKeyConstraint baseTableName=DEFAULT_CLIENT_SCOPE, constraintName=FK_R_DEF_CLI_SCOPE_SCOPE; dropForeignKeyConstraint baseTableName=CLIENT_SCOPE_CLIENT, constraintName=FK_C_CLI_SCOPE_SCOPE; dropForeignKeyConstraint baseTableName=CLIENT_SC...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('13.0.0-KEYCLOAK-17992-drop-constraints', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.091086', 93, 'MARK_RAN', '9:544d201116a0fcc5a5da0925fbbc3bde', 'dropPrimaryKey constraintName=C_CLI_SCOPE_BIND, tableName=CLIENT_SCOPE_CLIENT; dropIndex indexName=IDX_CLSCOPE_CL, tableName=CLIENT_SCOPE_CLIENT; dropIndex indexName=IDX_CL_CLSCOPE, tableName=CLIENT_SCOPE_CLIENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('13.0.0-increase-column-size-federated', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.100589', 94, 'EXECUTED', '9:43c0c1055b6761b4b3e89de76d612ccf', 'modifyDataType columnName=CLIENT_ID, tableName=CLIENT_SCOPE_CLIENT; modifyDataType columnName=SCOPE_ID, tableName=CLIENT_SCOPE_CLIENT', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('13.0.0-KEYCLOAK-17992-recreate-constraints', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.102446', 95, 'MARK_RAN', '9:8bd711fd0330f4fe980494ca43ab1139', 'addNotNullConstraint columnName=CLIENT_ID, tableName=CLIENT_SCOPE_CLIENT; addNotNullConstraint columnName=SCOPE_ID, tableName=CLIENT_SCOPE_CLIENT; addPrimaryKey constraintName=C_CLI_SCOPE_BIND, tableName=CLIENT_SCOPE_CLIENT; createIndex indexName=...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('json-string-accomodation-fixed', 'keycloak', 'META-INF/jpa-changelog-13.0.0.xml', '2025-02-25 13:12:13.10748', 96, 'EXECUTED', '9:e07d2bc0970c348bb06fb63b1f82ddbf', 'addColumn tableName=REALM_ATTRIBUTE; update tableName=REALM_ATTRIBUTE; dropColumn columnName=VALUE, tableName=REALM_ATTRIBUTE; renameColumn newColumnName=VALUE, oldColumnName=VALUE_NEW, tableName=REALM_ATTRIBUTE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('14.0.0-KEYCLOAK-11019', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.112381', 97, 'EXECUTED', '9:24fb8611e97f29989bea412aa38d12b7', 'createIndex indexName=IDX_OFFLINE_CSS_PRELOAD, tableName=OFFLINE_CLIENT_SESSION; createIndex indexName=IDX_OFFLINE_USS_BY_USER, tableName=OFFLINE_USER_SESSION; createIndex indexName=IDX_OFFLINE_USS_BY_USERSESS, tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('14.0.0-KEYCLOAK-18286', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.115871', 98, 'MARK_RAN', '9:259f89014ce2506ee84740cbf7163aa7', 'createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('14.0.0-KEYCLOAK-18286-revert', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.14344', 99, 'MARK_RAN', '9:04baaf56c116ed19951cbc2cca584022', 'dropIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('14.0.0-KEYCLOAK-18286-supported-dbs', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.147659', 100, 'EXECUTED', '9:60ca84a0f8c94ec8c3504a5a3bc88ee8', 'createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('14.0.0-KEYCLOAK-18286-unsupported-dbs', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.148661', 101, 'MARK_RAN', '9:d3d977031d431db16e2c181ce49d73e9', 'createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('KEYCLOAK-17267-add-index-to-user-attributes', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.152343', 102, 'EXECUTED', '9:0b305d8d1277f3a89a0a53a659ad274c', 'createIndex indexName=IDX_USER_ATTRIBUTE_NAME, tableName=USER_ATTRIBUTE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('KEYCLOAK-18146-add-saml-art-binding-identifier', 'keycloak', 'META-INF/jpa-changelog-14.0.0.xml', '2025-02-25 13:12:13.158072', 103, 'EXECUTED', '9:2c374ad2cdfe20e2905a84c8fac48460', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('15.0.0-KEYCLOAK-18467', 'keycloak', 'META-INF/jpa-changelog-15.0.0.xml', '2025-02-25 13:12:13.162426', 104, 'EXECUTED', '9:47a760639ac597360a8219f5b768b4de', 'addColumn tableName=REALM_LOCALIZATIONS; update tableName=REALM_LOCALIZATIONS; dropColumn columnName=TEXTS, tableName=REALM_LOCALIZATIONS; renameColumn newColumnName=TEXTS, oldColumnName=TEXTS_NEW, tableName=REALM_LOCALIZATIONS; addNotNullConstrai...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('17.0.0-9562', 'keycloak', 'META-INF/jpa-changelog-17.0.0.xml', '2025-02-25 13:12:13.165309', 105, 'EXECUTED', '9:a6272f0576727dd8cad2522335f5d99e', 'createIndex indexName=IDX_USER_SERVICE_ACCOUNT, tableName=USER_ENTITY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('18.0.0-10625-IDX_ADMIN_EVENT_TIME', 'keycloak', 'META-INF/jpa-changelog-18.0.0.xml', '2025-02-25 13:12:13.168208', 106, 'EXECUTED', '9:015479dbd691d9cc8669282f4828c41d', 'createIndex indexName=IDX_ADMIN_EVENT_TIME, tableName=ADMIN_EVENT_ENTITY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('18.0.15-30992-index-consent', 'keycloak', 'META-INF/jpa-changelog-18.0.15.xml', '2025-02-25 13:12:13.175981', 107, 'EXECUTED', '9:80071ede7a05604b1f4906f3bf3b00f0', 'createIndex indexName=IDX_USCONSENT_SCOPE_ID, tableName=USER_CONSENT_CLIENT_SCOPE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('19.0.0-10135', 'keycloak', 'META-INF/jpa-changelog-19.0.0.xml', '2025-02-25 13:12:13.193775', 108, 'EXECUTED', '9:9518e495fdd22f78ad6425cc30630221', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('20.0.0-12964-supported-dbs', 'keycloak', 'META-INF/jpa-changelog-20.0.0.xml', '2025-02-25 13:12:13.196911', 109, 'EXECUTED', '9:e5f243877199fd96bcc842f27a1656ac', 'createIndex indexName=IDX_GROUP_ATT_BY_NAME_VALUE, tableName=GROUP_ATTRIBUTE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('20.0.0-12964-unsupported-dbs', 'keycloak', 'META-INF/jpa-changelog-20.0.0.xml', '2025-02-25 13:12:13.197724', 110, 'MARK_RAN', '9:1a6fcaa85e20bdeae0a9ce49b41946a5', 'createIndex indexName=IDX_GROUP_ATT_BY_NAME_VALUE, tableName=GROUP_ATTRIBUTE', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('client-attributes-string-accomodation-fixed', 'keycloak', 'META-INF/jpa-changelog-20.0.0.xml', '2025-02-25 13:12:13.202206', 111, 'EXECUTED', '9:3f332e13e90739ed0c35b0b25b7822ca', 'addColumn tableName=CLIENT_ATTRIBUTES; update tableName=CLIENT_ATTRIBUTES; dropColumn columnName=VALUE, tableName=CLIENT_ATTRIBUTES; renameColumn newColumnName=VALUE, oldColumnName=VALUE_NEW, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('21.0.2-17277', 'keycloak', 'META-INF/jpa-changelog-21.0.2.xml', '2025-02-25 13:12:13.206773', 112, 'EXECUTED', '9:7ee1f7a3fb8f5588f171fb9a6ab623c0', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('21.1.0-19404', 'keycloak', 'META-INF/jpa-changelog-21.1.0.xml', '2025-02-25 13:12:13.224458', 113, 'EXECUTED', '9:3d7e830b52f33676b9d64f7f2b2ea634', 'modifyDataType columnName=DECISION_STRATEGY, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=LOGIC, tableName=RESOURCE_SERVER_POLICY; modifyDataType columnName=POLICY_ENFORCE_MODE, tableName=RESOURCE_SERVER', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('21.1.0-19404-2', 'keycloak', 'META-INF/jpa-changelog-21.1.0.xml', '2025-02-25 13:12:13.227536', 114, 'MARK_RAN', '9:627d032e3ef2c06c0e1f73d2ae25c26c', 'addColumn tableName=RESOURCE_SERVER_POLICY; update tableName=RESOURCE_SERVER_POLICY; dropColumn columnName=DECISION_STRATEGY, tableName=RESOURCE_SERVER_POLICY; renameColumn newColumnName=DECISION_STRATEGY, oldColumnName=DECISION_STRATEGY_NEW, tabl...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('22.0.0-17484-updated', 'keycloak', 'META-INF/jpa-changelog-22.0.0.xml', '2025-02-25 13:12:13.233126', 115, 'EXECUTED', '9:90af0bfd30cafc17b9f4d6eccd92b8b3', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('22.0.5-24031', 'keycloak', 'META-INF/jpa-changelog-22.0.0.xml', '2025-02-25 13:12:13.234176', 116, 'MARK_RAN', '9:a60d2d7b315ec2d3eba9e2f145f9df28', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('23.0.0-12062', 'keycloak', 'META-INF/jpa-changelog-23.0.0.xml', '2025-02-25 13:12:13.238827', 117, 'EXECUTED', '9:2168fbe728fec46ae9baf15bf80927b8', 'addColumn tableName=COMPONENT_CONFIG; update tableName=COMPONENT_CONFIG; dropColumn columnName=VALUE, tableName=COMPONENT_CONFIG; renameColumn newColumnName=VALUE, oldColumnName=VALUE_NEW, tableName=COMPONENT_CONFIG', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('23.0.0-17258', 'keycloak', 'META-INF/jpa-changelog-23.0.0.xml', '2025-02-25 13:12:13.240894', 118, 'EXECUTED', '9:36506d679a83bbfda85a27ea1864dca8', 'addColumn tableName=EVENT_ENTITY', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.0-9758', 'keycloak', 'META-INF/jpa-changelog-24.0.0.xml', '2025-02-25 13:12:13.248773', 119, 'EXECUTED', '9:502c557a5189f600f0f445a9b49ebbce', 'addColumn tableName=USER_ATTRIBUTE; addColumn tableName=FED_USER_ATTRIBUTE; createIndex indexName=USER_ATTR_LONG_VALUES, tableName=USER_ATTRIBUTE; createIndex indexName=FED_USER_ATTR_LONG_VALUES, tableName=FED_USER_ATTRIBUTE; createIndex indexName...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.0-9758-2', 'keycloak', 'META-INF/jpa-changelog-24.0.0.xml', '2025-02-25 13:12:13.253594', 120, 'EXECUTED', '9:bf0fdee10afdf597a987adbf291db7b2', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.0-26618-drop-index-if-present', 'keycloak', 'META-INF/jpa-changelog-24.0.0.xml', '2025-02-25 13:12:13.258653', 121, 'MARK_RAN', '9:04baaf56c116ed19951cbc2cca584022', 'dropIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.0-26618-reindex', 'keycloak', 'META-INF/jpa-changelog-24.0.0.xml', '2025-02-25 13:12:13.262098', 122, 'EXECUTED', '9:08707c0f0db1cef6b352db03a60edc7f', 'createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.2-27228', 'keycloak', 'META-INF/jpa-changelog-24.0.2.xml', '2025-02-25 13:12:13.266447', 123, 'EXECUTED', '9:eaee11f6b8aa25d2cc6a84fb86fc6238', 'customChange', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.2-27967-drop-index-if-present', 'keycloak', 'META-INF/jpa-changelog-24.0.2.xml', '2025-02-25 13:12:13.267478', 124, 'MARK_RAN', '9:04baaf56c116ed19951cbc2cca584022', 'dropIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('24.0.2-27967-reindex', 'keycloak', 'META-INF/jpa-changelog-24.0.2.xml', '2025-02-25 13:12:13.268798', 125, 'MARK_RAN', '9:d3d977031d431db16e2c181ce49d73e9', 'createIndex indexName=IDX_CLIENT_ATT_BY_NAME_VALUE, tableName=CLIENT_ATTRIBUTES', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-28265-tables', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.272779', 126, 'EXECUTED', '9:deda2df035df23388af95bbd36c17cef', 'addColumn tableName=OFFLINE_USER_SESSION; addColumn tableName=OFFLINE_CLIENT_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-28265-index-creation', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.275422', 127, 'EXECUTED', '9:3e96709818458ae49f3c679ae58d263a', 'createIndex indexName=IDX_OFFLINE_USS_BY_LAST_SESSION_REFRESH, tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-28265-index-cleanup', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.279852', 128, 'EXECUTED', '9:8c0cfa341a0474385b324f5c4b2dfcc1', 'dropIndex indexName=IDX_OFFLINE_USS_CREATEDON, tableName=OFFLINE_USER_SESSION; dropIndex indexName=IDX_OFFLINE_USS_PRELOAD, tableName=OFFLINE_USER_SESSION; dropIndex indexName=IDX_OFFLINE_USS_BY_USERSESS, tableName=OFFLINE_USER_SESSION; dropIndex ...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-28265-index-2-mysql', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.281103', 129, 'MARK_RAN', '9:b7ef76036d3126bb83c2423bf4d449d6', 'createIndex indexName=IDX_OFFLINE_USS_BY_BROKER_SESSION_ID, tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-28265-index-2-not-mysql', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.284207', 130, 'EXECUTED', '9:23396cf51ab8bc1ae6f0cac7f9f6fcf7', 'createIndex indexName=IDX_OFFLINE_USS_BY_BROKER_SESSION_ID, tableName=OFFLINE_USER_SESSION', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-org', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.292974', 131, 'EXECUTED', '9:5c859965c2c9b9c72136c360649af157', 'createTable tableName=ORG; addUniqueConstraint constraintName=UK_ORG_NAME, tableName=ORG; addUniqueConstraint constraintName=UK_ORG_GROUP, tableName=ORG; createTable tableName=ORG_DOMAIN', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('unique-consentuser', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.301519', 132, 'EXECUTED', '9:5857626a2ea8767e9a6c66bf3a2cb32f', 'customChange; dropUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHOGM8UEWRT, tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_LOCAL_CONSENT, tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_EXTERNAL_CONSENT, tableName=...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('unique-consentuser-mysql', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.302537', 133, 'MARK_RAN', '9:b79478aad5adaa1bc428e31563f55e8e', 'customChange; dropUniqueConstraint constraintName=UK_JKUWUVD56ONTGSUHOGM8UEWRT, tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_LOCAL_CONSENT, tableName=USER_CONSENT; addUniqueConstraint constraintName=UK_EXTERNAL_CONSENT, tableName=...', '', NULL, '4.25.1', NULL, NULL, '0489131228'),('25.0.0-28861-index-creation', 'keycloak', 'META-INF/jpa-changelog-25.0.0.xml', '2025-02-25 13:12:13.306617', 134, 'EXECUTED', '9:b9acb58ac958d9ada0fe12a5d4794ab1', 'createIndex indexName=IDX_PERM_TICKET_REQUESTER, tableName=RESOURCE_SERVER_PERM_TICKET; createIndex indexName=IDX_PERM_TICKET_OWNER, tableName=RESOURCE_SERVER_PERM_TICKET', '', NULL, '4.25.1', NULL, NULL, '0489131228');
INSERT INTO databasechangeloglock VALUES (1,false,NULL,NULL),(1000,false,NULL,NULL),(1001,false,NULL,NULL);
INSERT INTO default_client_scope VALUES ('master','13052fde-d239-4154-b80b-0f406ed76ded',false),('master','395ebcc0-2a2e-4f24-9f63-6d2cfeada3ab',false),('master','a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb',true),('master','abde17dd-48e0-4d26-a2b7-e75c04b1ac7f',false),('master','c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb',false),('master','cb25d275-eff3-4655-b032-e163a0a23c0f',true),('master','e4f019f4-8a8a-4682-bf50-8e883c89cd03',true),('master','f0e07760-3d3d-45d5-b651-403f8b19de35',true),('master','f55ceb89-6d3c-4bcb-882e-44c498d8b305',true);
INSERT INTO migration_model VALUES ('a59ul','15.0.2',1635795164);
INSERT INTO protocol_mapper VALUES ('036972dc-32e5-4370-bc20-7bd787304f91','zoneinfo','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('0cde626f-2b4f-4d62-8b0b-e632ae2f6dc0','audience resolve','openid-connect','oidc-audience-resolve-mapper','54905dd0-4ade-494e-9c35-ab2d445a99f5',NULL),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','upn','openid-connect','oidc-usermodel-property-mapper',NULL,'abde17dd-48e0-4d26-a2b7-e75c04b1ac7f'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','middle name','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','username','openid-connect','oidc-usermodel-property-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('2d044ac9-0899-47f7-86a1-f71fffd3f487','allowed web origins','openid-connect','oidc-allowed-origins-mapper',NULL,'e4f019f4-8a8a-4682-bf50-8e883c89cd03'),('467c2758-877f-44d4-ad79-d947350fc843','phone number','openid-connect','oidc-usermodel-attribute-mapper',NULL,'13052fde-d239-4154-b80b-0f406ed76ded'),('573c6159-a4f2-4767-957b-28bb898307d6','phone number verified','openid-connect','oidc-usermodel-attribute-mapper',NULL,'13052fde-d239-4154-b80b-0f406ed76ded'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','gender','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','website','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','role list','saml','saml-role-list-mapper',NULL,'f55ceb89-6d3c-4bcb-882e-44c498d8b305'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','given name','openid-connect','oidc-usermodel-property-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','family name','openid-connect','oidc-usermodel-property-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','picture','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('93113378-38c5-468c-ba6e-582a88eee4f3','realm roles','openid-connect','oidc-usermodel-realm-role-mapper',NULL,'f0e07760-3d3d-45d5-b651-403f8b19de35'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','locale','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','updated at','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('b9f9ef85-c17a-4f48-8d79-df663027d674','address','openid-connect','oidc-address-mapper',NULL,'c5b6705e-e0d8-48ec-8a01-7bdcb7ac2aeb'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','locale','openid-connect','oidc-usermodel-attribute-mapper','bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc',NULL),('bf71c52a-c4ac-4c40-884c-36eac49a5911','audience resolve','openid-connect','oidc-audience-resolve-mapper',NULL,'f0e07760-3d3d-45d5-b651-403f8b19de35'),('bfb8523f-4da7-4373-892c-8ad546873db9','groups','openid-connect','oidc-usermodel-realm-role-mapper',NULL,'abde17dd-48e0-4d26-a2b7-e75c04b1ac7f'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','birthdate','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('ca4cc8b3-6882-4216-acdd-b491bb219658','nickname','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('da471592-2af0-4299-b6a5-b71c9e48991b','profile','openid-connect','oidc-usermodel-attribute-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','full name','openid-connect','oidc-full-name-mapper',NULL,'a7f0db2f-bd6c-4b77-9b85-bd1a62c6dfeb'),('e85d523d-597b-4357-85ca-c3f8d98eb894','email verified','openid-connect','oidc-usermodel-property-mapper',NULL,'cb25d275-eff3-4655-b032-e163a0a23c0f'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','client roles','openid-connect','oidc-usermodel-client-role-mapper',NULL,'f0e07760-3d3d-45d5-b651-403f8b19de35'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','email','openid-connect','oidc-usermodel-property-mapper',NULL,'cb25d275-eff3-4655-b032-e163a0a23c0f');
INSERT INTO protocol_mapper_config VALUES ('036972dc-32e5-4370-bc20-7bd787304f91','true','access.token.claim'),('036972dc-32e5-4370-bc20-7bd787304f91','zoneinfo','claim.name'),('036972dc-32e5-4370-bc20-7bd787304f91','true','id.token.claim'),('036972dc-32e5-4370-bc20-7bd787304f91','String','jsonType.label'),('036972dc-32e5-4370-bc20-7bd787304f91','zoneinfo','user.attribute'),('036972dc-32e5-4370-bc20-7bd787304f91','true','userinfo.token.claim'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','true','access.token.claim'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','upn','claim.name'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','true','id.token.claim'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','String','jsonType.label'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','username','user.attribute'),('0e5b6089-e999-49d4-b8d5-4fac64fad1d9','true','userinfo.token.claim'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','true','access.token.claim'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','middle_name','claim.name'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','true','id.token.claim'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','String','jsonType.label'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','middleName','user.attribute'),('1a0f9afc-d076-4cb2-a0a0-201fe9bbeb0d','true','userinfo.token.claim'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','true','access.token.claim'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','preferred_username','claim.name'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','true','id.token.claim'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','String','jsonType.label'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','username','user.attribute'),('1f7b81ef-4446-4b48-9b9f-e4c0c7a81abd','true','userinfo.token.claim'),('467c2758-877f-44d4-ad79-d947350fc843','true','access.token.claim'),('467c2758-877f-44d4-ad79-d947350fc843','phone_number','claim.name'),('467c2758-877f-44d4-ad79-d947350fc843','true','id.token.claim'),('467c2758-877f-44d4-ad79-d947350fc843','String','jsonType.label'),('467c2758-877f-44d4-ad79-d947350fc843','phoneNumber','user.attribute'),('467c2758-877f-44d4-ad79-d947350fc843','true','userinfo.token.claim'),('573c6159-a4f2-4767-957b-28bb898307d6','true','access.token.claim'),('573c6159-a4f2-4767-957b-28bb898307d6','phone_number_verified','claim.name'),('573c6159-a4f2-4767-957b-28bb898307d6','true','id.token.claim'),('573c6159-a4f2-4767-957b-28bb898307d6','boolean','jsonType.label'),('573c6159-a4f2-4767-957b-28bb898307d6','phoneNumberVerified','user.attribute'),('573c6159-a4f2-4767-957b-28bb898307d6','true','userinfo.token.claim'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','true','access.token.claim'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','gender','claim.name'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','true','id.token.claim'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','String','jsonType.label'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','gender','user.attribute'),('58b9f316-ed7a-451c-b8cd-f2deb33b62f7','true','userinfo.token.claim'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','true','access.token.claim'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','website','claim.name'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','true','id.token.claim'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','String','jsonType.label'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','website','user.attribute'),('5abd0fe2-dfa1-4ce0-99b2-cd811e0c5df8','true','userinfo.token.claim'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','Role','attribute.name'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','Basic','attribute.nameformat'),('630d46bc-fe33-4462-a13d-0eb9031c1b6d','false','single'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','true','access.token.claim'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','given_name','claim.name'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','true','id.token.claim'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','String','jsonType.label'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','firstName','user.attribute'),('654a9342-21fa-47b3-93cf-42ee3dd56c0b','true','userinfo.token.claim'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','true','access.token.claim'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','family_name','claim.name'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','true','id.token.claim'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','String','jsonType.label'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','lastName','user.attribute'),('6b6a6347-32f4-4015-908b-6c1ae96bb533','true','userinfo.token.claim'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','true','access.token.claim'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','picture','claim.name'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','true','id.token.claim'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','String','jsonType.label'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','picture','user.attribute'),('8809bca8-02c2-45d3-8a84-9cc04fea1f2a','true','userinfo.token.claim'),('93113378-38c5-468c-ba6e-582a88eee4f3','true','access.token.claim'),('93113378-38c5-468c-ba6e-582a88eee4f3','realm_access.roles','claim.name'),('93113378-38c5-468c-ba6e-582a88eee4f3','String','jsonType.label'),('93113378-38c5-468c-ba6e-582a88eee4f3','true','multivalued'),('93113378-38c5-468c-ba6e-582a88eee4f3','foo','user.attribute'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','true','access.token.claim'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','locale','claim.name'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','true','id.token.claim'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','String','jsonType.label'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','locale','user.attribute'),('9fafbd6d-4ba4-4e9f-8346-acf7427d6597','true','userinfo.token.claim'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','true','access.token.claim'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','updated_at','claim.name'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','true','id.token.claim'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','String','jsonType.label'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','updatedAt','user.attribute'),('afda1f0b-3aa3-4559-8f94-f36276c9f3a2','true','userinfo.token.claim'),('b9f9ef85-c17a-4f48-8d79-df663027d674','true','access.token.claim'),('b9f9ef85-c17a-4f48-8d79-df663027d674','true','id.token.claim'),('b9f9ef85-c17a-4f48-8d79-df663027d674','country','user.attribute.country'),('b9f9ef85-c17a-4f48-8d79-df663027d674','formatted','user.attribute.formatted'),('b9f9ef85-c17a-4f48-8d79-df663027d674','locality','user.attribute.locality'),('b9f9ef85-c17a-4f48-8d79-df663027d674','postal_code','user.attribute.postal_code'),('b9f9ef85-c17a-4f48-8d79-df663027d674','region','user.attribute.region'),('b9f9ef85-c17a-4f48-8d79-df663027d674','street','user.attribute.street'),('b9f9ef85-c17a-4f48-8d79-df663027d674','true','userinfo.token.claim'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','true','access.token.claim'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','locale','claim.name'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','true','id.token.claim'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','String','jsonType.label'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','locale','user.attribute'),('bd9c1acb-36e7-46e4-b06f-9eee0caca2f3','true','userinfo.token.claim'),('bfb8523f-4da7-4373-892c-8ad546873db9','true','access.token.claim'),('bfb8523f-4da7-4373-892c-8ad546873db9','groups','claim.name'),('bfb8523f-4da7-4373-892c-8ad546873db9','true','id.token.claim'),('bfb8523f-4da7-4373-892c-8ad546873db9','String','jsonType.label'),('bfb8523f-4da7-4373-892c-8ad546873db9','true','multivalued'),('bfb8523f-4da7-4373-892c-8ad546873db9','foo','user.attribute'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','true','access.token.claim'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','birthdate','claim.name'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','true','id.token.claim'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','String','jsonType.label'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','birthdate','user.attribute'),('c6d317fc-5cca-43d0-8a37-f65018b9967e','true','userinfo.token.claim'),('ca4cc8b3-6882-4216-acdd-b491bb219658','true','access.token.claim'),('ca4cc8b3-6882-4216-acdd-b491bb219658','nickname','claim.name'),('ca4cc8b3-6882-4216-acdd-b491bb219658','true','id.token.claim'),('ca4cc8b3-6882-4216-acdd-b491bb219658','String','jsonType.label'),('ca4cc8b3-6882-4216-acdd-b491bb219658','nickname','user.attribute'),('ca4cc8b3-6882-4216-acdd-b491bb219658','true','userinfo.token.claim'),('da471592-2af0-4299-b6a5-b71c9e48991b','true','access.token.claim'),('da471592-2af0-4299-b6a5-b71c9e48991b','profile','claim.name'),('da471592-2af0-4299-b6a5-b71c9e48991b','true','id.token.claim'),('da471592-2af0-4299-b6a5-b71c9e48991b','String','jsonType.label'),('da471592-2af0-4299-b6a5-b71c9e48991b','profile','user.attribute'),('da471592-2af0-4299-b6a5-b71c9e48991b','true','userinfo.token.claim'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','true','access.token.claim'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','true','id.token.claim'),('e07d0a56-65b6-4a34-bc9c-124e9a85a344','true','userinfo.token.claim'),('e85d523d-597b-4357-85ca-c3f8d98eb894','true','access.token.claim'),('e85d523d-597b-4357-85ca-c3f8d98eb894','email_verified','claim.name'),('e85d523d-597b-4357-85ca-c3f8d98eb894','true','id.token.claim'),('e85d523d-597b-4357-85ca-c3f8d98eb894','boolean','jsonType.label'),('e85d523d-597b-4357-85ca-c3f8d98eb894','emailVerified','user.attribute'),('e85d523d-597b-4357-85ca-c3f8d98eb894','true','userinfo.token.claim'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','true','access.token.claim'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','resource_access.${client_id}.roles','claim.name'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','String','jsonType.label'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','true','multivalued'),('ec3d6c94-95ae-4baa-82c5-162893cc80d0','foo','user.attribute'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','true','access.token.claim'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','email','claim.name'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','true','id.token.claim'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','String','jsonType.label'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','email','user.attribute'),('fda628cc-59f0-4fc3-8197-627cc7012b2a','true','userinfo.token.claim');
INSERT INTO realm_attribute VALUES ('_browser_header.contentSecurityPolicy','master','frame-src ''self''; frame-ancestors ''self''; object-src ''none'';'),('_browser_header.contentSecurityPolicyReportOnly','master',''),('_browser_header.strictTransportSecurity','master','max-age=31536000; includeSubDomains'),('_browser_header.xContentTypeOptions','master','nosniff'),('_browser_header.xFrameOptions','master','SAMEORIGIN'),('_browser_header.xRobotsTag','master','none'),('_browser_header.xXSSProtection','master','1; mode=block'),('bruteForceProtected','master','false'),('defaultSignatureAlgorithm','master','RS256'),('displayName','master','Keycloak'),('displayNameHtml','master','<div class=\"kc-logo-text\"><span>Keycloak</span></div>'),('failureFactor','master','30'),('maxDeltaTimeSeconds','master','43200'),('maxFailureWaitSeconds','master','900'),('minimumQuickLoginWaitSeconds','master','60'),('offlineSessionMaxLifespan','master','5184000'),('offlineSessionMaxLifespanEnabled','master','false'),('permanentLockout','master','false'),('quickLoginCheckMilliSeconds','master','1000'),('waitIncrementSeconds','master','60');
INSERT INTO realm_events_listeners VALUES ('master','jboss-logging');
INSERT INTO realm_required_credential VALUES ('password','password',true,true,'master');
INSERT INTO redirect_uris VALUES ('4e4977d6-eaa9-4245-ae4c-04d20f5436d9','/realms/master/account/*'),('54905dd0-4ade-494e-9c35-ab2d445a99f5','/realms/master/account/*'),('5a059221-51fd-434f-84a6-40fa51cda5ce','https://app.localssl.dev/api/v1/oidc/redirect'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','/admin/master/console/*');
INSERT INTO required_action_provider VALUES ('0c0818a4-641d-42e0-9b86-8bfeb7ee7368','CONFIGURE_TOTP','Configure OTP','master',true,false,'CONFIGURE_TOTP',10),('1781b401-8336-4a8c-a102-4a092c723cd3','UPDATE_PASSWORD','Update Password','master',true,false,'UPDATE_PASSWORD',30),('1bbbf0d1-e6f8-42b4-8741-19d1c59af15f','delete_account','Delete Account','master',false,false,'delete_account',60),('49195d42-495c-42cf-828e-f736ff686b9b','update_user_locale','Update User Locale','master',true,false,'update_user_locale',1000),('4a942b81-ccbb-49f9-a510-db0dda0d4ed9','VERIFY_EMAIL','Verify Email','master',true,false,'VERIFY_EMAIL',50),('5508dda9-65dc-40de-9e49-baf5d918b980','terms_and_conditions','Terms and Conditions','master',false,false,'terms_and_conditions',20),('556d4b2e-61e8-40db-924f-b0f0bdcae242','UPDATE_PROFILE','Update Profile','master',true,false,'UPDATE_PROFILE',40);
INSERT INTO scope_mapping VALUES ('54905dd0-4ade-494e-9c35-ab2d445a99f5','1aa723ff-209c-4637-b3c8-8159d72e9b09');
INSERT INTO user_role_mapping VALUES ('4d227aaf-b4fa-4a86-9535-30210f612f2e','563bb06b-d712-48c1-9381-cd6473e18590'),('4d227aaf-b4fa-4a86-9535-30210f612f2e','744b396e-3cf9-4e9f-9493-61b7b188fb10'),('5827ab16-b5bc-4738-b05e-89406e065439','563bb06b-d712-48c1-9381-cd6473e18590');
INSERT INTO web_origins VALUES ('5a059221-51fd-434f-84a6-40fa51cda5ce','https://app.localssl.dev/'),('bda020f6-dd7f-4bb8-b565-bdc8edb9a8fc','+');