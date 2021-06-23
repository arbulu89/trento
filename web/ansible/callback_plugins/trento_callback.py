"""
Ansible callback plugin to update the consul service state at the end of the execution
and send the summary
"""

from ansible.plugins.callback import CallbackBase
import json
import socket
import paramiko
import logging

class CallbackModule(CallbackBase):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'aggregate'
    CALLBACK_NAME = 'trento'

    def __init__(self):
        self._hosts = {}
        super(CallbackModule, self).__init__()

    def v2_playbook_on_start(self, playbook):
        self._display.banner("Trento callback plugin started")

    def v2_runner_on_ok(self, result):
        host = result._host.get_name()
        for tag in result._task_fields.get('tags'):
            if tag.startswith("check:"):
                if host not in self._hosts:
                    self._hosts[host] = {"passed": 0, "failed": 0, "warning": 0}
                self._hosts[host]["passed"] += 1

    def v2_runner_on_failed(self, result, ignore_errors=False):
        status = "failed"
        check_found = False
        host = result._host.get_name()
        for tag in result._task_fields.get('tags'):
            if tag.startswith("check:"):
                check_found = True
                if host not in self._hosts:
                    self._hosts[host] = {"passed": 0, "failed": 0, "warning": 0}
            if tag == "on_failed:warning":
                status = "warning"
        if check_found:
            self._hosts[host][status] += 1

    def v2_playbook_on_stats(self, stats):
        cmds = []
        for host, value in self._hosts.items():
            output = "== Summary ==\n%d checks PASS\n%d checks WARN\n%d checks FAIL" % (value['passed'], value['warning'], value['failed'])
            host_ip = socket.gethostbyname(host)
            if value['failed']:
                status = "critical"
                data = json.dumps({"Status": "passing", "Output": output})
                #cmds.append("curl --request PUT --data '{}' http://{}:8500/v1/agent/check/update/ha_config_checker".format(data, host_ip))
            elif value['warning']:
                status = "warning"
                data = json.dumps({"Status": "passing", "Output": output})
                #cmds.append("curl --request PUT --data '{}' http://{}:8500/v1/agent/check/update/ha_config_checker".format(data, host_ip))
            else:
                status = "passing"
                data = json.dumps({"Status": "critical", "Output": output})
                #cmds.append("curl --request PUT --data '{}' http://{}:8500/v1/agent/check/update/ha_config_checker".format(data, host_ip))

            data = json.dumps({"Status": status, "Output": output})
            cmds.append("curl --request PUT --data '{}' http://{}:8500/v1/agent/check/update/ha_config_checker".format(data, host_ip))

            ssh = paramiko.SSHClient()
            ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
            ssh.load_system_host_keys()
            ssh.connect(host, port=22)
            for cmd in cmds:
                ssh.exec_command(cmd)
