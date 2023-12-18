#! /bin/python3
from sys import argv, exit
progname = argv.pop(0)
if not argv:
    print(f"usage:")
    print(f"$ {progname} [section or item]")
    exit(1)
reqs = argv.copy()

info = {
    "tipo":["paas", "saas", "iaas"],
    "serviço":["serviços de db (por demanda)", "armazenamento de banco de dados", "serviço de cache gerenciado de memória ram", "serviços de container", "serviço de computção sem servidor - serverless (por demanda)"],
    "sgbd":["mysql", "postgresql", "sqlserver", "oracle"],
    "vcpu":["4", "8", "16", "32"],
    "memória ram":["16gb", "32gb", "64gb", "128gb"],
    "memória cache":["mínimo 6gb", "mínimo 26gb", "mínimo 52gb"],
    "métrica":["instância/hora", "gb/mês", "unidade/hora", "milhão de requisições/mês", "gb/segundo"],
}

color_red = "\x1b[1;91m"
color_grey = "\x1b[1;90m"
color_nc = "\x1b[0m"

# sec_name: (sec_name in req, set[item in reqs])
out:dict[str, tuple[bool, set[str]]] = {}

for sec_name,items in info.items():
    for req_index in range(len(argv)):
        req = argv[req_index].lower()

        if req in sec_name:
            out[sec_name] = out.get(sec_name, (True, set()))
            out[sec_name] = (True, out[sec_name][1])

        for item_index in range(len(items)):
            if any([arg in items[item_index] for arg in argv]):
                out[sec_name] = out.get(sec_name, (False, set()))
                out[sec_name][1].add(items[item_index])

if not out:
    print(f"didn't find any item or section from \"{', '.join(reqs)}\"")

for sec in out:
    clr = color_red if out[sec][0] else color_nc
    print(f"{clr}{sec}{color_nc}:")
    for item_index in range(len(info[sec])):
        item = info[sec][item_index]
        clr = color_red if item in out[sec][1] else color_grey
        print(f"{clr}{item_index}{color_nc}:\t{clr}{item}{color_nc}")

