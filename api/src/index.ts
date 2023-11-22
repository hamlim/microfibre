/**
 * Welcome to Cloudflare Workers! This is your first worker.
 *
 * - Run `npm run dev` in your terminal to start a development server
 * - Open a browser tab at http://localhost:8787/ to see your worker in action
 * - Run `npm run deploy` to publish your worker
 *
 * Learn more at https://developers.cloudflare.com/workers/
 */

export interface Env {
	DB: D1Database;
	// Example binding to KV. Learn more at https://developers.cloudflare.com/workers/runtime-apis/kv/
	// MY_KV_NAMESPACE: KVNamespace;
	//
	// Example binding to Durable Object. Learn more at https://developers.cloudflare.com/workers/runtime-apis/durable-objects/
	// MY_DURABLE_OBJECT: DurableObjectNamespace;
	//
	// Example binding to R2. Learn more at https://developers.cloudflare.com/workers/runtime-apis/r2/
	// MY_BUCKET: R2Bucket;
	//
	// Example binding to a Service. Learn more at https://developers.cloudflare.com/workers/runtime-apis/service-bindings/
	// MY_SERVICE: Fetcher;
	//
	// Example binding to a Queue. Learn more at https://developers.cloudflare.com/queues/javascript-apis/
	// MY_QUEUE: Queue;
}

function protect(request: Request, env: Env) {
	return async function handler(cb: () => Promise<Response>): Promise<Response> {
		if (request.headers.get('x-auth-token') === 'yolo-swag') {
			return cb()
		}
		return new Response("Unauthenticated! Make sure you are providing the valid auth token!", {status: 403});
	}
}

async function handlePostRequest(request: Request, env: Env): Promise<Response> {
  try {
    const { body, location, created_timezone } = await request.json<{body: string; location: string; created_timezone: string;}>();

    // Insert the post into the posts table
    const {results} = await env.DB.prepare(`
      INSERT INTO posts (body, location, created_timezone, created_time)
      VALUES (?1, ?2, ?3, ?4)
    `).bind(body, location, created_timezone, Date.now()).all();

    // Get the last inserted row's ID
    const selectStatement = await env.DB.prepare(`
      SELECT last_insert_rowid() AS lastId
    `);
    const result = await selectStatement.first();
    const postId = result?.lastId;

    const responseJson = {
      status: 'success',
      postId: postId,
    };

    return new Response(JSON.stringify(responseJson), {
      status: 200,
      headers: new Headers({ 'content-type': 'application/json' }),
    });
  } catch (error) {
    const responseJson = {
      status: 'failure',
      error: (error as Error).message,
    };

    return new Response(JSON.stringify(responseJson), {
      status: 500,
      headers: new Headers({ 'content-type': 'application/json' }),
    });
  }
}

async function handlePostsRequest(request: Request, env: Env): Promise<Response> {
  try {
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page')) || 0;
    const pageSize = Number(url.searchParams.get('pageSize')) || 10;

    // Calculate the offset based on the page and pageSize
    const offset = page * pageSize;

    // Retrieve the posts from the posts table
    const selectPostsStatement = await env.DB.prepare(`
      SELECT *
      FROM posts
      ORDER BY id DESC
      LIMIT ?1 OFFSET ?2
    `).bind(pageSize, offset);
    const {results: posts} = await selectPostsStatement.all();

    // Retrieve the associated media entries for each post
    for (const post of posts) {
      const selectMediaStatement = await env.DB.prepare(`
        SELECT *
        FROM media
        WHERE post_id = ?1
      `).bind(post.id);
      const {results: mediaEntries} = await selectMediaStatement.all();
      post.media = mediaEntries;
    }

    const responseJson = {
      status: 'success',
      posts: posts,
    };

    return new Response(JSON.stringify(responseJson), {
      status: 200,
      headers: new Headers({ 'content-type': 'application/json' }),
    });
  } catch (error) {
    const responseJson = {
      status: 'failure',
      error: (error as Error).message,
    };

    return new Response(JSON.stringify(responseJson), {
      status: 500,
      headers: new Headers({ 'content-type': 'application/json' }),
    });
  }
}

async function handleUpdateRequest(request: Request, env: Env): Promise<Response> {
  try {
    const { postId, body, location, updated_timezone } = await request.json<{ postId: number; body?: string; location?: string; updated_timezone: string; }>();

		if (!postId || !updated_timezone) {
			return new Response(JSON.stringify({
				status: 'failure',
				error: 'Missing `postId` or `updated_timezone` in the payload!'
			}), {
        status: 500,
        headers: new Headers({ 'content-type': 'application/json' }),
      });
		}

    // Check if the post exists
    const selectStatement = await env.DB.prepare(`
      SELECT *
      FROM posts
      WHERE id = ?1
    `).bind(postId);
    const post = await selectStatement.first();

    if (!post) {
      const responseJson = {
        status: 'failure',
        error: 'Post not found',
      };

      return new Response(JSON.stringify(responseJson), {
        status: 404,
        headers: new Headers({ 'content-type': 'application/json' }),
      });
    }

    // Update the post
    const updateStatement = await env.DB.prepare(`
      UPDATE posts
      SET body = ?1, location = ?2, updated_timezone = ?3, updated_time = ?4
      WHERE id = ?5
    `).bind(body || post.body, location || post.location, updated_timezone, Date.now(), postId);
    await updateStatement.run();

    const responseJson = {
      status: 'success',
      postId: postId,
    };

    return new Response(JSON.stringify(responseJson), {
      status: 200,
      headers: new Headers({ 'content-type': 'application/json' }),
    });
  } catch (error) {
    const responseJson = {
      status: 'failure',
      error: (error as Error).message,
    };

    return new Response(JSON.stringify(responseJson), {
      status: 500,
      headers: new Headers({ 'content-type': 'application/json' }),
    });
  }
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		let {method} = request;
		let { pathname } = new URL(request.url);

		switch (method) {
			case 'POST': {
				switch (pathname) {
					case '/v1/post': {
						return protect(request, env)(() => handlePostRequest(request, env));
					}
					case '/v1/update': {
						return protect(request, env)(() => handleUpdateRequest(request, env))
					}
				}
			}
			case 'GET':
			default: {
				switch (pathname) {
					case '/v1/posts': {
						return protect(request, env)(() => handlePostsRequest(request, env));
					}
					case '/status': {
						let res, headers = new Headers();
						let accepts = request.headers.get('accept')?.split(',')
						if (accepts?.includes('text/html')) {
							res = `We're up! 🆙`
						} else {
							headers.append('content-type', 'application/json');
							res = JSON.stringify({
								status: 'up'
							})
						}
						return new Response(res, {
							status: 200,
							headers
						})
					}
					case '/':
					default: {
						let res, headers = new Headers();
						let accepts = request.headers.get('accept')?.split(',')
						if (accepts?.includes('text/html')) {
							headers.append('content-type', 'text/html');
							res = `<html><head><title>Microfibre v1 API</title></head><body><marquee>Hello World!</marquee></body></html>`
						} else {
							headers.append('content-type', 'application/json');
							res = JSON.stringify({
								status: 'up'
							})
						}
						return new Response(res, {
							status: 200,
							headers
						})
					}
				}
			}
		}
	},
};
