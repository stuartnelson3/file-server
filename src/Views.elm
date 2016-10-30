module Views exposing (..)


-- External Imports
import Html exposing (..)
import Html.App as Html
import Html.Attributes exposing (..)
import Html.Events exposing (..)


-- Internal Imports
import Types exposing (Model, ApiResponse, Msg, Route(..))


view : Model -> Html Msg
view model =
  case model.route of
    Movies ->
      moviesView model.movies

    Movie name ->
      let
        one = Debug.log "view: name" name
      in
        filmView model.movie

    _ ->
      notFoundView model


notFoundView : Model -> Html Msg
notFoundView model =
  div [] [
    h1 [] [ text "not found" ]
  ]


movieView : ApiResponse -> Html Msg
movieView model =
  a [ class "db link dim tc"
    , href ("#/movie/" ++ model.apiMovie.imdbID) ]
    [ filmView model ]


moviesView : List ApiResponse -> Html Msg
moviesView movies =
  div [
    classList [
      ("cf", True),
      ("pa2", True)
    ]
  ] (List.map movieView movies)


filmView : ApiResponse -> Html msg
filmView resp =
  let
      movie = resp.apiMovie
  in
    div [
        classList [
          ("fl", True),
          ("w-50", False),
          ("pa2", True),
          ("w-25-m", True),
          ("w-w-20-l", True)
        ]
      ]
      [ img [
          src movie.poster,
          classList [
            ("db", True),
            ("w-100", True),
            ("outline", True),
            ("black-10", True)
          ]
        ] []
      , dl [
          classList [
            ("mt2", True),
            ("f6", True),
            ("lh-copy", True)
            ]
        ] [
          filmMetaData movie.title,
          filmMetaData movie.year,
          filmMetaData movie.kind
        ]
      ]


filmMetaData : String -> Html msg
filmMetaData data =
  dt [ class "m10 black truncate w-100" ] [ text data ]

