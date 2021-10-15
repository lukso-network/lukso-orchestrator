#!/usr/bin/env bash
IP="127.0.0.45"
if [[ ! $IP =~ ^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$ ]]; then
  echo "Invalid IP"
fi



exit
DATADIR=.
if [[ ! -f $DATADIR/genesis-time.txt ]] || [[ $(cat $DATADIR/genesis-time.txt) -gt $(date +%s) ]]; then
  sed -i 's/--min-sync-peers=2/--disable-sync/g' $(which lukso);
fi

exit

echo $(cat genesis.txt) != $(date +%s)
if [[ $(cat genesis.txt) != $(date +%s) ]]; then
  echo looooooooooo
fi

echo -n $(date +%s) > genesis.txt




exit
pandora init $1
exit
PANDORA_VERBOSITY=trace

case $PANDORA_VERBOSITY in
  silent)
    PANDORA_VERBOSITY=0
    ;;
  error)
    PANDORA_VERBOSITY=1
    ;;
  trace|chuj)
    PANDORA_VERBOSITY=500
  esac

echo $PANDORA_VERBOSITY;
exit;

URL="https://raw.githubusercontent.com/lukso-network/lukso-orchestrator/feature/l15-setup-script/scripts/lukso"
STATUS_CODE=$(curl -s -o /dev/null -I -w "%{http_code}" $URL);
   if [[ $STATUS_CODE != "200" ]]; then
     echo "File not found, check URL: $URL" ;
     exit;
   fi
echo found

BOOTNODES=();

optspec=":hv-:"
while getopts "$optspec" optchar; do
    case "${optchar}" in
        h)
          echo "usage: $0 [-v] [--loglevel[=]<value>]" >&2
          exit 2
        ;;
        -)
            case "${OPTARG}" in

                bootnode)
                  val="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ));
                  BOOTNODES+=($val);
                ;;

                *)
                    if [ "$OPTERR" = 1 ] && [ "${optspec:0:1}" != ":" ]; then
                        echo "Unknown option --${OPTARG}" >&2
                    fi
                    ;;
            esac;;

        *)
            if [ "$OPTERR" != 1 ] || [ "${optspec:0:1}" = ":" ]; then
                echo "Non-option argument: '-${OPTARG}'" >&2
            fi
            ;;
    esac
done

echo ${BOOTNODES[@]};
