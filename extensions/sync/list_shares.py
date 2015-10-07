import utils

import api
import _sync as sync


def list_shares(data):
    syncthing = api.Syncthing(sync.SYNCTHING_URL)

    config = syncthing.config

    return config['folders']

if __name__ == '__main__':
    utils.run(list_shares)
