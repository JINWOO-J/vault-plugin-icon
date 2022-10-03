ui = true
#disable_sealwrap = true # for Enterprise version
disable_mlock = true

storage "raft" {
  path    = "./data"
  node_id = "node1"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_disable = "true"
}

path "sys/audit"
{
  capabilities = ["read", "sudo"]
}

#api_addr = "http://0.0.0.0:8200"
api_addr = "http://127.0.0.1:8200"
#cluster_addr = "https://20.20.5.136:8201"
#default_lease_ttl = "10s"
#max_lease_ttl = "10s"
plugin_directory= "./plugin"





#              plugin_directory= "/vault/file/plugin"
#              storage "raft" {
#                path    = "/vault/data"
#                node_id = "node1"
#              }
#              listener "tcp" {
#                address     = "0.0.0.0:8200"
#                cluster_address = "0.0.0.0:8201"
#                tls_disable = "true"
#              }
