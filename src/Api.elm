module Api exposing (..)

-- External Imports
import Http
import Json.Decode as Json exposing (..)
import Task
import String

-- Internal Imports
import Types exposing (..)


-- Api


baseUrl : String
baseUrl = "http://localhost:8080/api/v0"

searchApi : Cmd Msg
searchApi =
  let
      url = baseUrl ++ "/movies"
  in
     Task.perform FetchFail MoviesFetchSucceed (Http.get decodeApiResponse url)


singleMovieSearch : String -> Cmd Msg
singleMovieSearch id =
  let
      url = String.join "/" [baseUrl, "movie", id]
  in
     Task.perform FetchFail MovieFetchSucceed (Http.get responseDecoder url)


decodeApiResponse : Json.Decoder (List ApiResponse)
decodeApiResponse =
  Json.list responseDecoder


responseDecoder : Json.Decoder ApiResponse
responseDecoder =
  Json.object3 ApiResponse
    ("Title" := Json.string)
    ("FullPath" := Json.string)
    ("ApiMovie" := apiResponseDecoder)


apiResponseDecoder : Json.Decoder MovieResponse
apiResponseDecoder =
  Json.object5 MovieResponse
    ("ImdbID" := Json.string)
    ("Poster" := Json.string)
    ("Title" := Json.string)
    ("Type" := Json.string)
    ("Year" := Json.string)
