import { Settings } from "luxon";

Settings.defaultLocale = "en";
Settings.defaultZoneName = "UTC";

import clientConfig from "./config";
import { config } from "app/session";

config.setValues(clientConfig);

import MockAdapter from "axios-mock-adapter";
import Api from "common/api";

const Mock = new MockAdapter(Api, { onNoMatch: "throwException" });

const mockHeaders = {
  "Content-Type": "application/json; charset=utf-8",
};

const getCollectionResponse = [
  { id: 1, name: "John Smith" },
  { id: 1, name: "John Smith" },
];

const getEntityResponse = {
  id: 1,
  name: "John Smith",
};

const postEntityResponse = {
  users: [{ id: 1, name: "John Smith" }],
};

const putEntityResponse = {
  users: [{ id: 2, name: "John Foo" }],
};

const deleteEntityResponse = null;
Mock.onPost("api/v1/users/urii20d30w2wqzjf/profile").reply(
  200,
  { DisplayName: "Max New" },
  mockHeaders
);
Mock.onPost("api/v1/users/52/avatar").reply(200, { Thumb: "abc", ThumbSrc: "manual" }, mockHeaders);
Mock.onGet("api/v1/foo").reply(200, getCollectionResponse, mockHeaders);
Mock.onGet("api/v1/foo/123").reply(200, getEntityResponse, mockHeaders);
Mock.onPost("api/v1/foo").reply(201, postEntityResponse, mockHeaders);
Mock.onPut("api/v1/foo/2").reply(200, putEntityResponse, mockHeaders);
Mock.onDelete("api/v1/foo/2").reply(204, deleteEntityResponse, mockHeaders);
Mock.onGet("api/v1/error").reply(401, "custom error cat", mockHeaders);

Mock.onPost("api/v1/batch/photos/archive").reply(200, { photos: [1, 3] }, mockHeaders);
Mock.onPost("api/v1/photos/pqbemz8276mhtobh/approve").reply(200, {}, mockHeaders);
Mock.onPost("api/v1/photos/pqbemz8276mhtobh/files/fqbfk181n4ca5sud/primary").reply(
  200,
  {
    ID: 10,
    UID: "pqbemz8276mhtobh",
    Files: [
      {
        UID: "fqbfk181n4ca5sud",
        Name: "1980/01/superCuteKitten.mp4",
        Primary: true,
        FileType: "mp4",
        Hash: "1xxbgdt55",
      },
    ],
  },
  mockHeaders
);

Mock.onPut("api/v1/photos/pqbemz8276mhtobh").reply(
  200,
  {
    ID: 10,
    UID: "pqbemz8276mhtobh",
    TitleSrc: "manual",
    Files: [
      {
        UID: "fqbfk181n4ca5sud",
        Name: "1980/01/superCuteKitten.mp4",
        Primary: false,
        FileType: "mp4",
        Hash: "1xxbgdt55",
      },
    ],
  },
  mockHeaders
);

Mock.onDelete("api/v1/photos/abc123/unlike").reply(200);
Mock.onDelete("api/v1/photos/pqbemz8276mhtobh/files/fqbfk181n4ca5sud").reply(
  200,
  {
    success: "successfully deleted",
  },
  mockHeaders
);
Mock.onPost("api/v1/photos/pqbemz8276mhtobh/files/fqbfk181n4ca5sud/unstack").reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPost("api/v1/photos/pqbemz8276mhtobh/label", { Name: "Cat", Priority: 10 }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPut("api/v1/photos/pqbemz8276mhtobh/label/12345", { Uncertainty: 0 }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPut("api/v1/photos/pqbemz8276mhtobh/label/12345", { Label: { Name: "Sommer" } }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onDelete("api/v1/photos/pqbemz8276mhtobh/label/12345").reply(
  200,
  { success: "ok" },
  mockHeaders
);

Mock.onPost("api/v1/session").reply(
  200,
  {
    id: "5aa770f2a1ef431628d9f17bdf82a0d16865e99d4a1ddd9356e1aabfe6464683",
    access_token: "999900000000000000000000000000000000000000000000",
    provider: "test",
    data: { token: "123token" },
    user: { ID: 1, UID: "urjysof3b9v7lgex", Name: "test", Email: "test@test.com" },
  },
  mockHeaders
);

Mock.onGet("api/v1/session/a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f").reply(
  200,
  {
    id: "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f",
    access_token: "234200000000000000000000000000000000000000000000",
    provider: "public",
    data: { token: "123token" },
    user: { ID: 1, UID: "urjysof3b9v7lgex", Name: "test", Email: "test@test.com" },
  },
  mockHeaders
);

Mock.onGet("api/v1/session/5aa770f2a1ef431628d9f17bdf82a0d16865e99d4a1ddd9356e1aabfe6464683").reply(
  200,
  {
    id: "5aa770f2a1ef431628d9f17bdf82a0d16865e99d4a1ddd9356e1aabfe6464683",
    access_token: "999900000000000000000000000000000000000000000000",
    provider: "test",
    data: { token: "123token" },
    user: { ID: 1, UID: "urjysof3b9v7lgex", Name: "test", Email: "test@test.com" },
  },
  mockHeaders
);

Mock.onDelete(
  "api/v1/session/5aa770f2a1ef431628d9f17bdf82a0d16865e99d4a1ddd9356e1aabfe6464683"
).reply(200);

Mock.onDelete(
  "api/v1/session/a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f"
).reply(200);

Mock.onGet("api/v1/settings").reply(200, { download: true, language: "de" }, mockHeaders);
Mock.onPost("api/v1/settings").reply(200, { download: true, language: "en" }, mockHeaders);

Mock.onGet("api/v1/services/123/folders").reply(200, { foo: "folders" }, mockHeaders);
Mock.onPost("api/v1/services/123/upload").reply(200, { foo: "upload" }, mockHeaders);

Mock.onGet("api/v1/folders/2011/10-Halloween", {
  params: { recursive: true, uncached: true },
}).reply(
  200,
  { folders: [1, 2], files: [1] },
  {
    "Content-Type": "application/json; charset=utf-8",
    "x-count": "3",
    "x-limit": "100",
    "x-offset": "0",
  }
);
Mock.onGet("api/v1/folders/2011/10-Halloween", { params: { recursive: true } }).reply(
  200,
  {
    folders: [1, 2, 3],
    files: [1],
  },
  mockHeaders
);
Mock.onGet("api/v1/folders/originals/2011/10-Halloween", { params: { recursive: true } }).reply(
  200,
  {
    folders: [1, 2, 3],
    files: [1],
  },
  mockHeaders
);

Mock.onPut("albums/66/links/5").reply(
  200,
  {
    UID: 5,
    Slug: "friends",
    Expires: 80000,
    UpdatedAt: "2012-07-08T14:45:39Z",
  },
  mockHeaders
);

Mock.onGet("api/v1/albums/66").reply(200, { Success: "ok" }, mockHeaders);
Mock.onPost("api/v1/albums/66/links").reply(
  200,
  {
    Password: "passwd",
    Expires: 8000,
    Slug: "christmas-2019",
    Comment: "",
    Perm: 0,
  },
  mockHeaders
);
Mock.onDelete("api/v1/albums/66/links/5").reply(200, { Success: "ok" }, mockHeaders);
Mock.onGet("api/v1/albums/66/links").reply(
  200,
  [
    { UID: "sqcwh80ifesw74ht", ShareUID: "aqcwh7weohhk49q2", Slug: "july-2020" },
    { UID: "sqcwhxh1h58rf3c2", ShareUID: "aqcwh7weohhk49q2" },
  ],
  mockHeaders
);
Mock.onPut("/api/v1/albums/66").reply(
  200,
  {
    Description: "Test description",
  },
  mockHeaders
);

Mock.onGet("api/v1/albums").reply(
  200,
  {
    ID: 51,
    CreatedAt: "2019-07-03T18:48:07Z",
    UpdatedAt: "2019-07-25T01:04:44Z",
    DeletedAt: "0001-01-01T00:00:00Z",
    Slug: "tabby-cat",
    Name: "tabby cat",
    Priority: 5,
    LabelCount: 9,
    Favorite: false,
    Description: "",
    Notes: "",
  },
  {
    "Content-Type": "application/json; charset=utf-8",
    "x-count": "3",
    "x-limit": "100",
    "x-offset": "0",
  }
);

Mock.onOptions("api/v1/albums").reply(
  200,
  {
    foo: "bar",
  },
  mockHeaders
);
Mock.onOptions("api/v1/albums/abc").reply(
  200,
  {
    foo: "edit",
  },
  mockHeaders
);
Mock.onDelete("api/v1/albums/abc").reply(
  200,
  {
    status: "ok",
  },
  mockHeaders
);
Mock.onPut("api/v1/albums/abc").reply(
  200,
  {
    Description: "Test description",
  },
  mockHeaders
);

//Mock.onPost("api/v1/users/55/profile").reply(200, { DisplayName: "Max New" }, mockHeaders);
//Mock.onPost("users/55/profile").reply(200, { DisplayName: "Max New" }, mockHeaders);
//Mock.onPost("api/v1/users/55/profile").reply(200, { DisplayName: "Max New" }, mockHeaders);

Mock.onAny("api/v1/users/52/register").reply(200, { foo: "register" }, mockHeaders);

Mock.onAny("api/v1/users/53/profile").reply(200, { foo: "profile" }, mockHeaders);

Mock.onPut("api/v1/users/54/password").reply(
  200,
  { password: "old", new_password: "new" },
  mockHeaders
);

Mock.onGet("api/v1/link/5").reply(200, "get success", mockHeaders);
Mock.onPut("api/v1/link/5").reply(200, "put success", mockHeaders);
Mock.onDelete("api/v1/link/5").reply(200, "delete success", mockHeaders);

Mock.onPost("api/v1/photos/55/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onDelete("api/v1/photos/55/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onGet("api/v1/albums/5").reply(200, { UID: "5" }, mockHeaders);
Mock.onPut("api/v1/photos/5").reply(200, { UID: "5" }, mockHeaders);
Mock.onDelete("api/v1/photos/abc123/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onPost("api/v1/photos/5/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onPost("api/v1/labels/ABC123/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onDelete("api/v1/labels/ABC123/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onPost("api/v1/folders/dqbevau2zlhxrxww/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onDelete("api/v1/folders/dqbevau2zlhxrxww/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onPost("api/v1/photos/undefined/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onDelete("api/v1/photos/undefined/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onPost("api/v1/albums/5/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onDelete("api/v1/albums/5/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onGet("api/v1/config").reply(200, clientConfig, mockHeaders);
Mock.onPut("api/v1/markers/mBC123ghytr", { Review: false, Invalid: false }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPut("api/v1/markers/mCC123ghytr", { Review: false, Invalid: true }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPut("api/v1/markers/mDC123ghytr", { SubjSrc: "manual", Name: "testname" }).reply(
  200,
  {
    success: "ok",
    Name: "testname",
  },
  mockHeaders
);
Mock.onDelete("api/v1/markers/mEC123ghytr/subject").reply(200, { success: "ok" }, mockHeaders);
Mock.onPut("api/v1/faces/f123ghytrfggd", { Hidden: false }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPut("api/v1/faces/f123ghytrfggd", { Hidden: true }).reply(
  200,
  {
    success: "ok",
  },
  mockHeaders
);
Mock.onPost("api/v1/subjects/s123ghytrfggd/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onPut("api/v1/subjects/s123ghytrfggd").reply(200, { status: "ok" }, mockHeaders);
Mock.onDelete("api/v1/subjects/s123ghytrfggd/like").reply(200, { status: "ok" }, mockHeaders);
Mock.onGet("api/v1/config/options").reply(200, { success: "ok" }, mockHeaders);
Mock.onPost("api/v1/config/options").reply(200, { success: "ok" }, mockHeaders);
Mock.onPost("api/v1/albums").reply(200, { success: "ok" }, mockHeaders);

//Mock.onPost().reply(200);
//Mock.onDelete().reply(200);
/*
Mock.onPost().reply(200).onDelete().reply(200);
Mock.onDelete().reply(200);
Mock.onAny().reply(200, "editForm");
Mock.onPut().reply(200, { Description: "Test description" });
Mock.onPut().reply(200, { Description: "Test description" });
Mock.onPost().reply(200, { Description: "Test description" });
*/

export { Api, Mock };
