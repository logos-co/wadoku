DEFAULT_CONFIG_FILE = "github.com/logos-co/wadoku/waku/config.json"
DEFAULT_RUN_PAIR = "lf"


def get_config_file(args):
    return DEFAULT_CONFIG_FILE if not hasattr(args, "config") else args.config


def get_run_pair(args):
    return DEFAULT_RUN_PAIR if not hasattr(args, "run_pair") else args.run_pair


def run(args):
    print(args)
    config_file = get_config_file(args)
    run_pair = get_run_pair(args)
    print("Reading the config from: %s" %config_file)
    print("The runpair is %s" %run_pair)
    config_json = read_file(src=config_file)
    config = json.decode(config_json)

    #input_file = config['input_file']
    output_file = config['output_file']
    duration = config['duration']
    iat = config['iat']
    loglvl = config['log_level']
    ctopic = config['content_topic']
    print(config)


    if run_pair == "lf":                # run lightpush and filter

      waku_filtr = add_service(
        service_id = "waku-filter",
        config = struct(
            image = "waku-filter:alpha",
            entrypoint= ["/go/bin/waku-filter"],
            cmd = [ "-o=" + "/go/bin/out/filter.out",
                    "-d=" + duration,
                    "-c=" + ctopic,
                    "-l=" + loglvl,
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
                    "-c=" + ctopic,
                    "-l=" + loglvl,
                    "-i=" + iat ],
        ),
      )
      print(waku_filtr, waku_lpush)

    else:                             # run waku publish and subscribe

      waku_sub = add_service(
        service_id = "waku-subscribe",
        config = struct(
            image = "waku-subscribe:alpha",
            entrypoint= ["/go/bin/waku-subscribe"],
            cmd = [ "-o=" + "/go/bin/out/subscribe.out",
                    "-d=" + duration,
                    "-c=" + ctopic,
                    "-l=" + loglvl,
                    "-i=" + iat ],
        ),
     )
      waku_pub = add_service(
        service_id = "waku-publish",
        config = struct(
            image = "waku-publish:alpha",
            entrypoint= ["/go/bin/waku-publish"],
            cmd = [ "-o=" + "/go/bin/out/publish.out",
                    "-d=" + duration,
                    "-c=" + ctopic,
                    "-l=" + loglvl,
                    "-i=" + iat ],
        ),
      )
      print(waku_sub, waku_pub)
