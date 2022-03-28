/*
 * Flow Go SDK
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"context"
	"fmt"

	"github.com/onflow/flow-go/engine/access/rest/models"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client/convert"

	"github.com/onflow/flow-go-sdk"
)

const SEALED_HEIGHT = "sealed"
const EMULATOR_API = "http://127.0.0.1:8888/v1"
const TESTNET_API = "https://rest-testnet.onflow.org/v1/"
const MAINNET_API = "https://rest-mainnet.onflow.org/v1/"
const CANARYNET_API = ""

type handler interface {
	getBlockByID(ctx context.Context, ID string) (*models.Block, error)
	getBlockByHeight(ctx context.Context, height string) ([]*models.Block, error)
	getAccount(ctx context.Context, address string, height string) (*models.Account, error)
	getCollection(ctx context.Context, ID string) (*models.Collection, error)
	executeScriptAtBlockHeight(ctx context.Context, height string, script string, arguments []string) (string, error)
	executeScriptAtBlockID(ctx context.Context, ID string, script string, arguments []string) (string, error)
	getTransaction(ctx context.Context, ID string, includeResult bool) (*models.Transaction, error)
	sendTransaction(ctx context.Context, transaction []byte) error
	getEvents(ctx context.Context, eventType string, start string, end string, blockIDs []string) ([]models.BlockEvents, error)
}

// NewClient creates an instance of the client with the provided http handler.
func NewClient(handler handler) *Client {
	return &Client{handler}
}

// NewDefaultEmulatorClient creates a new client for connecting to the emulator AN API.
func NewDefaultEmulatorClient(debug bool) (*Client, error) {
	httpHandler, err := newHandler(EMULATOR_API, debug)
	if err != nil {
		return nil, err
	}

	return NewClient(httpHandler), nil
}

// NewDefaultTestnetClient creates a new client for connecting to the testnet AN API.
func NewDefaultTestnetClient() (*Client, error) {
	httpHandler, err := newHandler(TESTNET_API, false)
	if err != nil {
		return nil, err
	}

	return NewClient(httpHandler), nil
}

// NewDefaultCanaryClient creates a new client for connecting to the canary AN API.
func NewDefaultCanaryClient() (*Client, error) {
	httpHandler, err := newHandler(CANARYNET_API, false)
	if err != nil {
		return nil, err
	}

	return NewClient(httpHandler), nil
}

// NewDefaultMainnetClient creates a new client for connecting to the mainnet AN API.
func NewDefaultMainnetClient() (*Client, error) {
	httpHandler, err := newHandler(MAINNET_API, false)
	if err != nil {
		return nil, err
	}

	return NewClient(httpHandler), nil
}

// Client implementing all the network interactions according to the client interface.
type Client struct {
	handler handler
}

func (c *Client) Ping(ctx context.Context) error {
	panic("implement me")
}

func (c *Client) GetBlockByID(ctx context.Context, blockID flow.Identifier) (*flow.Block, error) {
	block, err := c.handler.getBlockByID(ctx, blockID.String())
	if err != nil {
		return nil, err
	}

	return convert.HTTPToBlock(block)
}

func (c *Client) GetLatestBlockHeader(ctx context.Context, isSealed bool) (*flow.BlockHeader, error) {
	block, err := c.GetLatestBlock(ctx, isSealed)
	if err != nil {
		return nil, err
	}

	return &block.BlockHeader, nil
}

func (c *Client) GetBlockHeaderByID(ctx context.Context, blockID flow.Identifier) (*flow.BlockHeader, error) {
	block, err := c.GetBlockByID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	return &block.BlockHeader, nil
}

func (c *Client) GetBlockHeaderByHeight(ctx context.Context, height uint64) (*flow.BlockHeader, error) {
	block, err := c.GetBlockByHeight(ctx, height)
	if err != nil {
		return nil, err
	}

	return &block.BlockHeader, nil
}

func (c *Client) GetLatestBlock(ctx context.Context, isSealed bool) (*flow.Block, error) {
	blocks, err := c.handler.getBlockByHeight(ctx, convert.SealedToHTTP(isSealed))
	if err != nil {
		return nil, err
	}

	return convert.HTTPToBlock(blocks[0])
}

func (c *Client) GetBlockByHeight(ctx context.Context, height uint64) (*flow.Block, error) {
	blocks, err := c.handler.getBlockByHeight(ctx, fmt.Sprintf("%d", height))
	if err != nil {
		return nil, err
	}

	return convert.HTTPToBlock(blocks[0])
}

func (c *Client) GetCollection(ctx context.Context, ID flow.Identifier) (*flow.Collection, error) {
	collection, err := c.handler.getCollection(ctx, ID.String())
	if err != nil {
		return nil, err
	}

	return convert.HTTPToCollection(collection), nil
}

func (c *Client) SendTransaction(ctx context.Context, tx flow.Transaction) error {
	convertedTx, err := convert.TransactionToHTTP(tx)
	if err != nil {
		return err
	}

	return c.handler.sendTransaction(ctx, convertedTx)
}

func (c *Client) GetTransaction(ctx context.Context, ID flow.Identifier) (*flow.Transaction, error) {
	tx, err := c.handler.getTransaction(ctx, ID.String(), false)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToTransaction(tx)
}

func (c *Client) GetTransactionResult(ctx context.Context, ID flow.Identifier) (*flow.TransactionResult, error) {
	tx, err := c.handler.getTransaction(ctx, ID.String(), true)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToTransactionResult(tx.Result)
}

func (c *Client) GetAccount(ctx context.Context, address flow.Address) (*flow.Account, error) {
	account, err := c.handler.getAccount(ctx, address.String(), SEALED_HEIGHT)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToAccount(account)
}

func (c *Client) GetAccountAtLatestBlock(ctx context.Context, address flow.Address) (*flow.Account, error) {
	return c.GetAccount(ctx, address)
}

func (c *Client) GetAccountAtBlockHeight(
	ctx context.Context,
	address flow.Address,
	blockHeight uint64,
) (*flow.Account, error) {
	account, err := c.handler.getAccount(ctx, address.String(), fmt.Sprintf("%d", blockHeight))
	if err != nil {
		return nil, err
	}

	return convert.HTTPToAccount(account)
}

func (c *Client) ExecuteScriptAtLatestBlock(
	ctx context.Context,
	script []byte,
	arguments []cadence.Value,
) (cadence.Value, error) {
	args, err := convert.CadenceArgsToHTTP(arguments)
	if err != nil {
		return nil, err
	}

	result, err := c.handler.executeScriptAtBlockHeight(ctx, SEALED_HEIGHT, convert.ScriptToHTTP(script), args)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToCadenceValue(result)
}

func (c *Client) ExecuteScriptAtBlockID(
	ctx context.Context,
	blockID flow.Identifier,
	script []byte,
	arguments []cadence.Value,
) (cadence.Value, error) {
	args, err := convert.CadenceArgsToHTTP(arguments)
	if err != nil {
		return nil, err
	}

	result, err := c.handler.executeScriptAtBlockID(ctx, blockID.String(), convert.ScriptToHTTP(script), args)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToCadenceValue(result)
}

func (c *Client) ExecuteScriptAtBlockHeight(
	ctx context.Context,
	height uint64,
	script []byte,
	arguments []cadence.Value,
) (cadence.Value, error) {
	args, err := convert.CadenceArgsToHTTP(arguments)
	if err != nil {
		return nil, err
	}

	result, err := c.handler.executeScriptAtBlockHeight(ctx, fmt.Sprintf("%d", height), convert.ScriptToHTTP(script), args)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToCadenceValue(result)
}

func (c *Client) GetEventsForHeightRange(
	ctx context.Context,
	eventType string,
	startHeight uint64,
	endHeight uint64,
) ([]flow.BlockEvents, error) {
	events, err := c.handler.getEvents(
		ctx,
		eventType,
		fmt.Sprintf("%d", startHeight),
		fmt.Sprintf("%d", endHeight),
		nil,
	)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToBlockEvents(events)
}

func (c *Client) GetEventsForBlockIDs(
	ctx context.Context,
	eventType string,
	blockIDs []flow.Identifier,
) ([]flow.BlockEvents, error) {
	ids := make([]string, len(blockIDs))
	for i, id := range blockIDs {
		ids[i] = id.String()
	}

	events, err := c.handler.getEvents(ctx, eventType, "", "", ids)
	if err != nil {
		return nil, err
	}

	return convert.HTTPToBlockEvents(events)
}

func (c *Client) GetLatestProtocolStateSnapshot(ctx context.Context) ([]byte, error) {
	panic("implement me")
}

func (c *Client) GetExecutionResultForBlockID(ctx context.Context, blockID flow.Identifier) (*flow.ExecutionResult, error) {
	panic("implement me")
}
