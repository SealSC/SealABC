/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package hotStuff

import (
    "SealABC/crypto/hashes"
    "SealABC/crypto/signers"
    "SealABC/network"
    "time"
)

type Config struct {
    //signer and members
    SelfSigner signers.ISigner

    //member list
    Members []Member

    //timers config
    MemberOnlineCheckInterval   time.Duration
    ConsensusTimeout            time.Duration

    //new consensus round interval
    ConsensusInterval time.Duration

    //network
    Network network.Service

    //crypto
    SingerGenerator signers.ISignerGenerator
    HashCalc        hashes.IHashCalculator
}
