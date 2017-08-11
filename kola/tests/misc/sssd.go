// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package misc

import (
	"time"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform/conf"
	"github.com/coreos/mantle/util"
)

func init() {
	register.Register(&register.Test{
		Run:         SssdId,
		ClusterSize: 1,
		Name:        "coreos.auth.sssd.id",
		// This will not work on qemu, since it's pulling a docker image over
		// the netwok
		ExcludePlatforms: []string{"qemu"},
		UserData: conf.ContainerLinuxConfig(`systemd:
  units:
    - name: "sssd.service"
      enable: true
    - name: "openldap.service"
      enable: true
      contents: |
          [Unit]
          Description=Openldap Container
          After=docker.service
          Requires=docker.service
           
          [Service]
          TimeoutStartSec=0
          Restart=always
          ExecStart=/usr/bin/docker run -p 127.0.0.1:389:389 -p 127.0.0.1:389:389/udp --name openldap \
            -v /opt/openldap/ldif/pminsky.ldif:/container/service/slapd/assets/config/bootstrap/ldif/custom/pminsky.ldif:Z \
            -v /opt/openldap/ldif/acls.ldif:/root/acls.ldif:Z \
            osixia/openldap:1.1.9 --copy-service
           
          [Install]
          WantedBy=multi-user.target
storage:
  files:
    - filesystem: "root"
      path: "/opt/openldap/ldif/acls.ldif"
      mode: 0644
      contents:
        inline: |
            dn: olcDatabase={-1}frontend,cn=config
            changetype: modify
            replace: olcAccess
            olcAccess: {1}to dn.base="" by * read
            
            dn: olcDatabase={1}hdb,cn=config
            changetype: modify
            replace: olcAccess
            olcAccess: {1}to * by self write by dn="cn=admin,dc=example,dc=org" write by * read
    - filesystem: "root"
      path: "/opt/openldap/ldif/pminsky.ldif"
      mode: 0644
      contents:
        inline: |
            dn: ou=People,dc=example,dc=org
            changetype: add
            objectclass: top
            objectclass: organizationalUnit
            ou: People
            
            dn: cn=Pete Minsky,ou=People,dc=example,dc=org
            changetype: add
            objectclass: top
            objectclass: person
            objectclass: organizationalPerson
            objectclass: inetOrgPerson
            objectclass: posixAccount
            cn: Pete Minsky
            givenName: Pete
            sn: Minsky
            ou: People
            uid: pminsky
            uidNumber: 600
            gidNumber: 600
            homeDirectory: /home/pminsky
            userpassword: foo
    - filesystem: "root"
      path: "/etc/sssd/sssd.conf"
      mode: 0600
      contents:
        inline: |
          [sssd]
          config_file_version = 2
          services = nss, pam
          domains = LDAP
          
          [nss]
          
          [pam]
          
          [domain/LDAP]
          id_provider = ldap
          auth_provider = ldap
          ldap_default_bind_dn = cn=admin,dc=example,dc=org
          ldap_default_authtok_type = password
          ldap_default_authtok = admin
          ldap_schema = rfc2307
          ldap_uri = ldap://localhost
          ldap_search_base = dc=example,dc=org`),
	})
}

func SssdId(c cluster.TestCluster) {
	m := c.Machines()[0]

	var ldapmodifySucceeded bool
	var output []byte

	// Because openldap is run in a docker container that needs to be fetched
	// over the network, ldap won't be running when the machine starts and until
	// some time after that. Perform the following actions in a loop every 30
	// seconds to account for this:
	// - change the ACLs via `docker exec <name> ldapmodify` to allow anonymous
	//   binds to read everything. This is necessary for sssd to be able to look
	//   up user information
	// - run `id pminsky` to see if the `pminsky` user exists. If it does, then
	//   sssd successfully looked up a user from ldap.
	//
	// If the test runs for more than 5 minutes, fail.
	err := util.Retry(10, time.Second*30, func() error {
		var err error
		if !ldapmodifySucceeded {
			// This cannot be seeded via the same method that the pminsky user
			// is injected into ldap because it must run under different
			// credentials (config's creds) to be able to modify the ACLs, so
			// it's run manually.
			output, err = m.SSH(`docker exec openldap ldapmodify -x -H ldap://localhost  -D "cn=admin,cn=config" -w config -f /root/acls.ldif`)
			if err == nil {
				ldapmodifySucceeded = true
			}
		}
		if ldapmodifySucceeded {
			output, err = m.SSH("id pminsky")
			if err == nil {
				output, err = m.SSH("getent passwd pminsky")
				if err == nil {
					return nil
				}
			}
		}
		return err
	})
	if err != nil {
		c.Fatalf("test timed out, last command gave output %q and error %v", output, err)
	}
}
