DEFAULT_CONFIG_FILE = "github.com/logos-co/wadoku/waku/config.json"

def get_config_file(args):
    return DEFAULT_CONFIG_FILE if not hasattr(args, "config") else args.config


def run(args):
    print(args)
    config_file = get_config_file(args)
    print("Reading the config from: %s" %config_file)
    config_json = read_file(src=config_file)
    config = json.decode(config_json)

    #input_file = config['input_file']
    output_file = config['output_file']
    duration = config['duration']
    iat = config['iat']
    mount_src = config['mount_src']
    mount_dst = config['mount_target']
    print(config)

    waku_filtr = add_service(
        service_id = "waku-filter",
        config = struct(
            image = "waku-filter:alpha",
            entrypoint= ["/go/bin/waku-filter"],
            cmd = [ "-o=" + "/go/bin/out/filter.out",
                    "-d=" + duration,
                    "-i=" + iat ],
        ),
    )
    waku_lpush = add_service(
        service_id = "waku-lightpush",
        config = struct(
            image = "waku-lightpush:alpha",
            entrypoint= ["/go/bin/waku-lightpush"],
            cmd = [ "-o=" + "/go/bin/out/lightpush.out",
                    "-d=" + duration,
                    "-i=" + iat ],
        ),
    )

    print(waku_filtr, waku_lpush)

