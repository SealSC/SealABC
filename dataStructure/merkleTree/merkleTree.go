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

package merkleTree

import (
    "github.com/SealSC/SealABC/crypto/hashes/sha3"
    "bytes"
    "errors"
    "github.com/cbergoon/merkletree"
)

type content struct {
    hash []byte
}

func (c content)CalculateHash() ([]byte, error) {
    return c.hash, nil
}

func (c content) Equals(other merkletree.Content) (bool, error) {
    return bytes.Equal(c.hash, other.(content).hash), nil
}

type Tree struct {
    list []merkletree.Content
    tree *merkletree.MerkleTree
}

func (t *Tree) AddData(data [][]byte)  {
    for _, d := range data {
        hash := sha3.Sha256.Sum(d)
        t.AddHash(hash)
    }
}

func (t *Tree) AddHash(hash []byte)  {
    newContent := content{
        hash: hash,
    }

    t.list = append(t.list, newContent)
}

func (t *Tree) Calculate() (root [] byte, err error) {
    if len(t.list) == 0 {
        return
    }

    tree, err := merkletree.NewTreeWithHashStrategy(t.list, sha3.Sha256.OriginalHash())
    if err != nil {
        return
    }

    root = tree.MerkleRoot()

    return
}

func (t Tree) Verify() (passed bool, err error)  {
    if t.tree == nil {
        return false, errors.New("not calculated")
    }
    return t.tree.VerifyTree()
}

func (t Tree) VerifyData(data []byte) (passed bool, err error) {
    return t.VerifyHash(sha3.Sha256.Sum(data))
}

func (t Tree) VerifyHash(hash []byte) (passed bool, err error) {
    if t.tree == nil {
        return false, errors.New("not calculated")
    }

    newContent := content{
        hash: hash,
    }
    return t.tree.VerifyContent(newContent)
}