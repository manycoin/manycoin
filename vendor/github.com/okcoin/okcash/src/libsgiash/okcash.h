/*
  This file is part of okcash.

  okcash is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  okcash is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with okcash.  If not, see <http://www.gnu.org/licenses/>.
*/

/** @file okcash.h
* @date 2015
*/
#pragma once

#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stddef.h>
#include "compiler.h"

#define OKCASH_REVISION 23
#define OKCASH_DATASET_BYTES_INIT 1073741824U // 2**30
#define OKCASH_DATASET_BYTES_GROWTH 8388608U  // 2**23
#define OKCASH_CACHE_BYTES_INIT 1073741824U // 2**24
#define OKCASH_CACHE_BYTES_GROWTH 131072U  // 2**17
#define OKCASH_EPOCH_LENGTH 30000U
#define OKCASH_MIX_BYTES 128
#define OKCASH_HASH_BYTES 64
#define OKCASH_DATASET_PARENTS 256
#define OKCASH_CACHE_ROUNDS 3
#define OKCASH_ACCESSES 64
#define OKCASH_DAG_MAGIC_NUM_SIZE 8
#define OKCASH_DAG_MAGIC_NUM 0xFEE1DEADBADDCAFE

#ifdef __cplusplus
extern "C" {
#endif

/// Type of a seedhash/blockhash e.t.c.
typedef struct okcash_h256 { uint8_t b[32]; } okcash_h256_t;

// convenience macro to statically initialize an h256_t
// usage:
// okcash_h256_t a = okcash_h256_static_init(1, 2, 3, ... )
// have to provide all 32 values. If you don't provide all the rest
// will simply be unitialized (not guranteed to be 0)
#define okcash_h256_static_init(...)			\
	{ {__VA_ARGS__} }

struct okcash_light;
typedef struct okcash_light* okcash_light_t;
struct okcash_full;
typedef struct okcash_full* okcash_full_t;
typedef int(*okcash_callback_t)(unsigned);

typedef struct okcash_return_value {
	okcash_h256_t result;
	okcash_h256_t mix_hash;
	bool success;
} okcash_return_value_t;

/**
 * Allocate and initialize a new okcash_light handler
 *
 * @param block_number   The block number for which to create the handler
 * @return               Newly allocated okcash_light handler or NULL in case of
 *                       ERRNOMEM or invalid parameters used for @ref okcash_compute_cache_nodes()
 */
okcash_light_t okcash_light_new(uint64_t block_number);
/**
 * Frees a previously allocated okcash_light handler
 * @param light        The light handler to free
 */
void okcash_light_delete(okcash_light_t light);
/**
 * Calculate the light client data
 *
 * @param light          The light client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               an object of okcash_return_value_t holding the return values
 */
okcash_return_value_t okcash_light_compute(
	okcash_light_t light,
	okcash_h256_t const header_hash,
	uint64_t nonce
);

/**
 * Allocate and initialize a new okcash_full handler
 *
 * @param light         The light handler containing the cache.
 * @param callback      A callback function with signature of @ref okcash_callback_t
 *                      It accepts an unsigned with which a progress of DAG calculation
 *                      can be displayed. If all goes well the callback should return 0.
 *                      If a non-zero value is returned then DAG generation will stop.
 *                      Be advised. A progress value of 100 means that DAG creation is
 *                      almost complete and that this function will soon return succesfully.
 *                      It does not mean that the function has already had a succesfull return.
 * @return              Newly allocated okcash_full handler or NULL in case of
 *                      ERRNOMEM or invalid parameters used for @ref okcash_compute_full_data()
 */
okcash_full_t okcash_full_new(okcash_light_t light, okcash_callback_t callback);

/**
 * Frees a previously allocated okcash_full handler
 * @param full    The light handler to free
 */
void okcash_full_delete(okcash_full_t full);
/**
 * Calculate the full client data
 *
 * @param full           The full client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               An object of okcash_return_value to hold the return value
 */
okcash_return_value_t okcash_full_compute(
	okcash_full_t full,
	okcash_h256_t const header_hash,
	uint64_t nonce
);
/**
 * Get a pointer to the full DAG data
 */
void const* okcash_full_dag(okcash_full_t full);
/**
 * Get the size of the DAG data
 */
uint64_t okcash_full_dag_size(okcash_full_t full);

/**
 * Calculate the seedhash for a given block number
 */
okcash_h256_t okcash_get_seedhash(uint64_t block_number);

#ifdef __cplusplus
}
#endif
